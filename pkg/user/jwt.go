package user

import (
	"context"
	"errors"
	"time"

	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	"github.com/byteflowing/go-common/idx"
	"github.com/golang-jwt/jwt/v5"
)

type JwtService struct {
	issuer     string
	secretKey  string
	signMethod jwt.SigningMethod
	accessTTL  time.Duration
	refreshTTL time.Duration
	repo       Repo
	limiter    Limiter
	blk        BlockList
}

type JwtOpts struct {
	Issuer     string
	SecretKey  string
	SignMethod jwt.SigningMethod
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Repo       Repo
	Limiter    Limiter
	BlockList  BlockList
}

func NewJwtService(opts *JwtOpts) *JwtService {
	return &JwtService{
		issuer:     opts.Issuer,
		secretKey:  opts.SecretKey,
		signMethod: opts.SignMethod,
		accessTTL:  opts.AccessTTL,
		refreshTTL: opts.RefreshTTL,
		repo:       opts.Repo,
		limiter:    opts.Limiter,
		blk:        opts.BlockList,
	}
}

func (s *JwtService) GenerateToken(ctx context.Context, req *GenerateJwtReq) (accessToken, refreshToken string, err error) {
	if err = s.limiter.Allow(ctx, req.UserBasic.ID, enumsv1.UserLimitType_USER_LIMIT_TYPE_SIGN_IN); err != nil {
		return "", "", err
	}
	accessToken, refreshToken, accessClaims, refreshClaims, err := s.generateJwtToken(ctx, req)
	if err != nil {
		return "", "", err
	}
	if err = s.repo.AddSignInLog(ctx, req.SignInReq, accessClaims, refreshClaims); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (s *JwtService) VerifyAccessToken(ctx context.Context, tokenStr string) (claims *JwtClaims, err error) {
	return s.verifyToken(ctx, tokenStr, enumsv1.TokenType_TOKEN_TYPE_ACCESS)
}

func (s *JwtService) VerifyRefreshToken(ctx context.Context, tokenStr string) (claims *JwtClaims, err error) {
	return s.verifyToken(ctx, tokenStr, enumsv1.TokenType_TOKEN_TYPE_REFRESH)
}

func (s *JwtService) ExpireAccessToken(ctx context.Context, accessSessionID string) (err error) {
	log, err := s.repo.GetSignInLogByAccess(ctx, accessSessionID)
	if err != nil {
		return err
	}
	claims := &JwtClaims{
		Uid:  log.UID,
		Type: int32(enumsv1.TokenType_TOKEN_TYPE_ACCESS),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.UnixMilli(log.AccessExpiredAt)),
			ID:        log.AccessSessionID,
		},
	}
	if err = s.revokeToken(ctx, claims); err != nil {
		return err
	}
	return s.repo.ExpireAccessSessionID(ctx, accessSessionID)
}

func (s *JwtService) RefreshToken(ctx context.Context, req *GenerateJwtReq) (newAccessToken, newRefreshToken string, err error) {
	if err = s.limiter.Allow(ctx, req.UserBasic.ID, enumsv1.UserLimitType_USER_LIMIT_TYPE_REFRESH); err != nil {
		return "", "", err
	}
	var claims []*JwtClaims
	refreshClaim, err := s.VerifyRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return "", "", err
	}
	claims = append(claims, refreshClaim)
	log, err := s.repo.GetSignInLogByRefresh(ctx, refreshClaim.ID)
	if err != nil {
		return "", "", err
	}
	// 销毁token只关心下面填充的数据
	claims = append(claims, &JwtClaims{
		Uid:  log.UID,
		Type: int32(enumsv1.TokenType_TOKEN_TYPE_ACCESS),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.UnixMilli(log.AccessExpiredAt)),
			ID:        log.AccessSessionID,
		},
	})
	if err = s.revokeTokens(ctx, claims); err != nil {
		return "", "", err
	}
	newAccessToken, newRefreshToken, newAccessClaims, newRefreshClaims, err := s.generateJwtToken(ctx, req)
	if err != nil {
		return "", "", err
	}
	if err = s.repo.RefreshSignInLog(ctx, refreshClaim.ID, newAccessClaims, newRefreshClaims); err != nil {
		return "", "", err
	}
	return newAccessToken, newRefreshToken, nil
}

func (s *JwtService) RevokeByAccessToken(ctx context.Context, tokenStr string, status SessionStatus) error {
	accessClaims, err := s.parseClaims(tokenStr, false)
	if err != nil {
		return err
	}
	if accessClaims.Type != int32(enumsv1.TokenType_TOKEN_TYPE_ACCESS) {
		return errors.New("invalid token type")
	}
	log, err := s.repo.GetSignInLogByAccess(ctx, accessClaims.ID)
	if err != nil {
		return err
	}
	willRevokeClaims := []*JwtClaims{
		accessClaims,
		&JwtClaims{
			Uid:  log.UID,
			Type: int32(enumsv1.TokenType_TOKEN_TYPE_REFRESH),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.UnixMilli(log.AccessExpiredAt)),
				ID:        log.RefreshSessionID,
			},
		},
	}
	if err = s.revokeTokens(ctx, willRevokeClaims); err != nil {
		return err
	}
	return s.repo.DisActiveSignInLogByAccess(ctx, accessClaims.ID, status)
}

// RevokeTokens 同时撤销access和refresh token
// 必须成对出现，最多两个
func (s *JwtService) RevokeTokens(ctx context.Context, tokens []string) error {
	if len(tokens) != 2 {
		return errors.New("tokens length must be 2")
	}
	var claims []*JwtClaims
	for _, token := range tokens {
		claim, err := s.parseClaims(token, false)
		if err != nil {
			return err
		}
		claims = append(claims, claim)
	}
	return s.revokeTokens(ctx, claims)
}

func (s *JwtService) generateJwtToken(ctx context.Context, req *GenerateJwtReq) (accessToken, refreshToken string, accessClaims, refreshClaims *JwtClaims, err error) {
	now := time.Now()
	accessClaims = s.createClaims(now, enumsv1.TokenType_TOKEN_TYPE_ACCESS, req)
	refreshClaims = s.createClaims(now, enumsv1.TokenType_TOKEN_TYPE_REFRESH, req)
	accessToken, err = s.createToken(accessClaims)
	if err != nil {
		return
	}
	refreshToken, err = s.createToken(refreshClaims)
	if err != nil {
		return
	}
	return
}

func (s *JwtService) verifyToken(ctx context.Context, token string, t enumsv1.TokenType) (*JwtClaims, error) {
	parsedClaims, err := s.parseClaims(token, true)
	if err != nil {
		return nil, err
	}
	if parsedClaims.Type != int32(t) {
		return nil, ecode.ErrUserTokenMisMatch
	}
	// 验证token是否在黑名单中
	isBlock, err := s.blk.Exists(ctx, parsedClaims.ID)
	if err != nil {
		return nil, err
	}
	if isBlock {
		return nil, ecode.ErrUserTokenRevoked
	}
	return parsedClaims, nil
}

func (s *JwtService) revokeToken(ctx context.Context, claims *JwtClaims) error {
	ttl := s.ttlFromClaims(claims)
	if ttl <= 0 {
		return nil
	}
	return s.blk.Add(ctx, claims.ID, ttl)
}

func (s *JwtService) revokeTokens(ctx context.Context, items []*JwtClaims) error {
	var sessions []*SessionItem
	for _, item := range items {
		ttl := s.ttlFromClaims(item)
		if ttl <= 0 {
			continue
		}
		sessions = append(sessions, &SessionItem{
			SessionID: item.ID,
			TTL:       ttl,
		})
	}
	if len(sessions) > 0 {
		return s.blk.BatchAdd(ctx, sessions)
	}
	return nil
}

func (s *JwtService) parseClaims(tokenStr string, validate bool) (*JwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JwtClaims{}, s.keyFunc)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JwtClaims)
	if !ok {
		return nil, ecode.ErrUserInvalidToken
	}
	if validate {
		if claims.ExpiresAt.Unix() <= time.Now().Unix() {
			return nil, ecode.ErrUserTokenExpired
		}
		if !token.Valid {
			return nil, ecode.ErrUserInvalidToken
		}
	}
	return claims, nil
}

func (s *JwtService) keyFunc(token *jwt.Token) (interface{}, error) {
	if token.Method != s.signMethod {
		return nil, ecode.ErrUserTokenMisMatch
	}
	return []byte(s.secretKey), nil
}

func (s *JwtService) ttlFromClaims(claims *JwtClaims) time.Duration {
	if claims.ExpiresAt == nil {
		return 0
	}
	exp := claims.ExpiresAt.Unix() - time.Now().Unix()
	if exp <= 0 {
		return 0
	}
	return time.Duration(exp) * time.Second
}

func (s *JwtService) createClaims(t time.Time, typ enumsv1.TokenType, req *GenerateJwtReq) *JwtClaims {
	return &JwtClaims{
		Uid:      req.UserBasic.ID,
		Type:     int32(typ),
		AuthType: int32(req.AuthType),
		AppId:    req.AppId,
		OpenId:   req.OpenId,
		UnionId:  req.UnionId,
		Extra:    req.ExtraJwtClaims,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   req.UserBasic.Number,
			ExpiresAt: jwt.NewNumericDate(t.Add(s.accessTTL)),
			NotBefore: jwt.NewNumericDate(t),
			IssuedAt:  jwt.NewNumericDate(t),
			ID:        idx.UUIDv4(),
		},
	}
}

func (s *JwtService) createToken(claims *JwtClaims) (string, error) {
	accessJwt := jwt.NewWithClaims(s.signMethod, claims)
	return accessJwt.SignedString([]byte(s.secretKey))
}
