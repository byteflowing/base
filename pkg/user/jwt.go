package user

import (
	"context"
	"errors"
	"time"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/go-common/idx"
	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type JwtClaims struct {
	Uid   uint64
	Type  TokenType
	Extra interface{}
	jwt.RegisteredClaims
}

type JwtService struct {
	issuer     string
	secretKey  string
	signMethod jwt.SigningMethod
	accessTTL  time.Duration
	refreshTTL time.Duration
	repo       Repo
	bkl        BlockList
}

type JwtOpts struct {
	Issuer     string
	SecretKey  string
	SignMethod jwt.SigningMethod
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Repo       Repo
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
		bkl:        opts.BlockList,
	}
}

func (s *JwtService) GenerateToken(ctx context.Context, userBasic *model.UserBasic, req *SignInReq) (accessToken, refreshToken string, err error) {
	if err = s.bkl.CheckTokenLimit(ctx, uint64(userBasic.ID), LimitTypeSignIn); err != nil {
		return "", "", err
	}
	accessToken, refreshToken, accessClaims, refreshClaims, err := s.generateJwtToken(ctx, userBasic, req.ExtraJwtClaims)
	if err != nil {
		return "", "", err
	}
	if err = s.repo.AddSignInLog(ctx, req, accessClaims, refreshClaims); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (s *JwtService) VerifyAccessToken(ctx context.Context, tokenStr string) (claims *JwtClaims, err error) {
	return s.verifyToken(ctx, tokenStr, TokenTypeAccess)
}

func (s *JwtService) VerifyRefreshToken(ctx context.Context, tokenStr string) (claims *JwtClaims, err error) {
	return s.verifyToken(ctx, tokenStr, TokenTypeRefresh)
}

func (s *JwtService) ExpireAccessToken(ctx context.Context, accessSessionID string) (err error) {
	log, err := s.repo.GetSignInLogByAccess(ctx, accessSessionID)
	if err != nil {
		return err
	}
	claims := &JwtClaims{
		Uid:  uint64(log.UID),
		Type: TokenTypeAccess,
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

func (s *JwtService) RefreshToken(ctx context.Context, refreshToken string, userBasic *model.UserBasic, extraClaims map[string]any) (newAccessToken, newRefreshToken string, err error) {
	if err = s.bkl.CheckTokenLimit(ctx, uint64(userBasic.ID), LimitTypeSignIn); err != nil {
		return "", "", err
	}
	var claims []*JwtClaims
	refreshClaim, err := s.VerifyRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}
	claims = append(claims, refreshClaim)
	log, err := s.repo.GetSignInLogByRefresh(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}
	// 销毁token只关心下面填充的数据
	claims = append(claims, &JwtClaims{
		Uid:  uint64(log.UID),
		Type: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.UnixMilli(log.AccessExpiredAt)),
			ID:        log.AccessSessionID,
		},
	})
	if err = s.revokeTokens(ctx, claims); err != nil {
		return "", "", err
	}
	newAccessToken, newRefreshToken, newAccessClaims, newRefreshClaims, err := s.generateJwtToken(ctx, userBasic, extraClaims)
	if err != nil {
		return "", "", err
	}
	if err = s.repo.RefreshSignInLog(ctx, refreshToken, newAccessClaims, newRefreshClaims); err != nil {
		return "", "", err
	}
	return newAccessToken, newRefreshToken, nil
}

func (s *JwtService) RevokeByAccessToken(ctx context.Context, tokenStr string, status SessionStatus) error {
	accessClaims, err := s.parseClaims(tokenStr, false)
	if err != nil {
		return err
	}
	if accessClaims.Type != TokenTypeAccess {
		return errors.New("invalid token type")
	}
	log, err := s.repo.GetSignInLogByAccess(ctx, accessClaims.ID)
	if err != nil {
		return err
	}
	willRevokeClaims := []*JwtClaims{
		accessClaims,
		&JwtClaims{
			Uid:  uint64(log.UID),
			Type: TokenTypeRefresh,
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

func (s *JwtService) generateJwtToken(ctx context.Context, userBasic *model.UserBasic, extraClaims interface{}) (accessToken, refreshToken string, accessClaims, refreshClaims *JwtClaims, err error) {
	now := time.Now()
	accessClaims = &JwtClaims{
		Uid:   uint64(userBasic.ID),
		Extra: extraClaims,
		Type:  TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userBasic.Name,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        idx.UUIDv4(),
		},
	}
	refreshClaims = &JwtClaims{
		Uid:   uint64(userBasic.ID),
		Extra: extraClaims,
		Type:  TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userBasic.Name,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        idx.UUIDv4(),
		},
	}
	accessJwt := jwt.NewWithClaims(s.signMethod, accessClaims)
	accessToken, err = accessJwt.SignedString([]byte(s.secretKey))
	if err != nil {
		return
	}
	refreshJwt := jwt.NewWithClaims(s.signMethod, refreshClaims)
	refreshToken, err = refreshJwt.SignedString([]byte(s.secretKey))
	if err != nil {
		return
	}
	return
}

func (s *JwtService) verifyToken(ctx context.Context, token string, t TokenType) (*JwtClaims, error) {
	parsedClaims, err := s.parseClaims(token, true)
	if err != nil {
		return nil, err
	}
	if parsedClaims.Type != t {
		return nil, ecode.ErrUserTokenMisMatch
	}
	// 验证token是否在黑名单中
	isBlock, err := s.bkl.Exists(ctx, parsedClaims.ID)
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
	return s.bkl.Add(ctx, claims.ID, ttl)
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
		return s.bkl.BatchAdd(ctx, sessions)
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
