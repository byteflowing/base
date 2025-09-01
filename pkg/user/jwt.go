package user

import (
	"context"
	"time"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	"github.com/byteflowing/go-common/idx"
	"github.com/golang-jwt/jwt/v5"
)

type JwtService struct {
	issuer            string
	secretKey         string
	signMethod        jwt.SigningMethod
	accessTTL         time.Duration
	refreshTTL        time.Duration
	repo              Repo
	limiter           Limiter
	blk               BlockList
	maxActiveSessions int
}

type JwtOpts struct {
	Issuer            string
	SecretKey         string
	SignMethod        jwt.SigningMethod
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	Repo              Repo
	Limiter           Limiter
	BlockList         BlockList
	MaxActiveSessions int
}

func NewJwtService(opts *JwtOpts) *JwtService {
	return &JwtService{
		issuer:            opts.Issuer,
		secretKey:         opts.SecretKey,
		signMethod:        opts.SignMethod,
		accessTTL:         opts.AccessTTL,
		refreshTTL:        opts.RefreshTTL,
		repo:              opts.Repo,
		limiter:           opts.Limiter,
		blk:               opts.BlockList,
		maxActiveSessions: opts.MaxActiveSessions,
	}
}

func (s *JwtService) GenerateToken(ctx context.Context, req *GenerateJwtReq) (accessToken, refreshToken string, accessClaims, freshClaims *JwtClaims, err error) {
	if err = s.limiter.Allow(ctx, req.UserBasic.ID, enumsv1.UserLimitType_USER_LIMIT_TYPE_SIGN_IN); err != nil {
		return "", "", nil, nil, err
	}
	// 检查最大同时在线数
	logs, err := s.repo.GetActiveSignInLogs(ctx, req.UserBasic.ID)
	if err != nil {
		return "", "", nil, nil, err
	}
	if len(logs) >= s.maxActiveSessions {
		log := logs[len(logs)-1]
		if err = s.RevokeByLog(ctx, log); err != nil {
			return "", "", nil, nil, err
		}
		log.Status = int16(enumsv1.SessionStatus_SESSION_STATUS_KICKED_OUT)
		if err = s.repo.UpdateSignInLogByID(ctx, log); err != nil {
			return "", "", nil, nil, err
		}
	}
	accessToken, refreshToken, accessClaims, refreshClaims, err := s.generateJwtToken(ctx, req)
	if err != nil {
		return "", "", nil, nil, err
	}
	if err = s.repo.AddSignInLog(ctx, req.SignInReq, accessClaims, refreshClaims); err != nil {
		return "", "", nil, nil, err
	}
	return
}

func (s *JwtService) VerifyAccessToken(ctx context.Context, tokenStr string) (claims *JwtClaims, err error) {
	return s.verifyToken(ctx, tokenStr, enumsv1.TokenType_TOKEN_TYPE_ACCESS)
}

func (s *JwtService) VerifyRefreshToken(ctx context.Context, tokenStr string) (claims *JwtClaims, err error) {
	return s.verifyToken(ctx, tokenStr, enumsv1.TokenType_TOKEN_TYPE_REFRESH)
}

func (s *JwtService) RevokeTokens(ctx context.Context, items []*SessionItem) (err error) {
	return s.revoke(ctx, items)
}

func (s *JwtService) RevokeByLog(ctx context.Context, log *model.UserSignLog) (err error) {
	items := []*SessionItem{
		{
			SessionID: log.AccessSessionID,
			TTL:       s.TTLFromExpiredAt(log.AccessExpiredAt),
		},
		{
			SessionID: log.RefreshSessionID,
			TTL:       s.TTLFromExpiredAt(log.RefreshExpiredAt),
		},
	}
	return s.revoke(ctx, items)
}

func (s *JwtService) RefreshToken(ctx context.Context, req *GenerateJwtReq) (newAccessToken, newRefreshToken string, newAccessClaims, newRefreshClaims *JwtClaims, err error) {
	if err = s.limiter.Allow(ctx, req.UserBasic.ID, enumsv1.UserLimitType_USER_LIMIT_TYPE_REFRESH); err != nil {
		return "", "", nil, nil, err
	}
	refreshClaim, err := s.VerifyRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return "", "", nil, nil, err
	}
	log, err := s.repo.GetSignInLogByRefresh(ctx, refreshClaim.ID)
	if err != nil {
		return "", "", nil, nil, err
	}
	if err = s.RevokeByLog(ctx, log); err != nil {
		return "", "", nil, nil, err
	}
	newAccessToken, newRefreshToken, newAccessClaims, newRefreshClaims, err = s.generateJwtToken(ctx, req)
	if err != nil {
		return "", "", nil, nil, err
	}
	log.AccessSessionID = newAccessClaims.ID
	log.RefreshSessionID = newRefreshClaims.ID
	log.AccessExpiredAt = newAccessClaims.ExpiresAt.Unix()
	log.RefreshExpiredAt = newRefreshClaims.ExpiresAt.Unix()
	if err = s.repo.UpdateSignInLogByID(ctx, log); err != nil {
		return "", "", nil, nil, err
	}
	return
}

func (s *JwtService) TTLFromExpiredAt(expiredAt int64) time.Duration {
	if expiredAt <= 0 {
		return 0
	}
	exp := time.Unix(expiredAt, 0).Unix() - time.Now().Unix()
	if exp <= 0 {
		return 0
	}
	return time.Duration(exp) * time.Second
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

func (s *JwtService) revoke(ctx context.Context, items []*SessionItem) error {
	var willAdd []*SessionItem
	for _, item := range items {
		if item.TTL > 0 {
			willAdd = append(willAdd, item)
		}
	}
	if len(willAdd) == 0 {
		return nil
	}
	if len(willAdd) > 1 {
		return s.blk.BatchAdd(ctx, willAdd)
	}
	return s.blk.Add(ctx, willAdd[0].SessionID, willAdd[0].TTL)
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
