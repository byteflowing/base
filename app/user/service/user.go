package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/byteflowing/base/app/global_id"
	"github.com/byteflowing/base/app/user/auth"
	"github.com/byteflowing/base/app/user/auth/huawei"
	"github.com/byteflowing/base/app/user/auth/tencent"
	"github.com/byteflowing/base/app/user/common"
	"github.com/byteflowing/base/app/user/dal/model"
	"github.com/byteflowing/base/app/user/dal/query"
	"github.com/byteflowing/base/app/user/migrate"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/blocklist"
	"github.com/byteflowing/base/pkg/jwt"
	"github.com/byteflowing/base/pkg/utils/trans"
	"github.com/byteflowing/base/singleton"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	userv1 "github.com/byteflowing/proto/gen/go/user/v1"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type UserService struct {
	authProviders map[enumsv1.SignInType]auth.Auth
	db            *query.Query
	blk           *blocklist.BlockList
	token         *jwt.Jwt
	cfg           *userv1.UserConfig
	userv1.UnimplementedUserServiceServer
}

func NewUserService(cfg *configv1.Config) *UserService {
	authProviders := newAuthProvider(cfg)
	orm := singleton.NewDB(cfg.Db)
	rdb := singleton.NewRDB(cfg.Redis)
	db := query.Use(orm)
	blk := blocklist.NewBlockList(cfg.User.KeyPrefix, rdb)
	token := jwt.New(cfg.User.Jwt.Issuer, cfg.User.Jwt.SecretKey)
	u := &UserService{
		authProviders: authProviders,
		db:            db,
		blk:           blk,
		token:         token,
		cfg:           cfg.User,
	}
	if cfg.User.AutoMigrate {
		m := migrate.NewMigrate(orm)
		if err := m.MigrateDB(); err != nil {
			panic(err)
		}
	}
	return u
}

func (u *UserService) SignUp(ctx context.Context, req *userv1.SignUpReq) (*userv1.SignUpResp, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserService) SignIn(ctx context.Context, req *userv1.SignInReq) (*userv1.SignInResp, error) {
	provider, err := u.getAuthProvider(req.SignInType)
	if err != nil {
		return nil, err
	}
	var resp *userv1.SignInResp
	err = u.db.Transaction(func(tx *query.Query) error {
		result, err := provider.Authenticate(ctx, req, tx)
		if err != nil {
			return err
		}
		user := result.User
		accessToken, refreshToken, err := u.genToken(user, req.ExtraJwtClaims)
		agent := req.Agent
		if agent == nil {
			agent = &userv1.Agent{}
		}
		if err := tx.UserSignLog.WithContext(ctx).Create(&model.UserSignLog{
			TenantID:         user.GetTenantId(),
			UID:              user.GetUid(),
			Type:             int16(req.SignInType),
			Status:           int16(user.GetStatus()),
			Identifier:       result.Identifier,
			IP:               agent.Ip,
			Location:         common.LocationToString(agent.Location),
			Agent:            agent.Agent,
			Device:           agent.Device,
			AccessJti:        accessToken.Jti,
			RefreshJti:       refreshToken.Jti,
			AccessExpiredAt:  trans.Ref(accessToken.Exp),
			RefreshExpiredAt: trans.Ref(refreshToken.Exp),
		}); err != nil {
			return err
		}
		resp = &userv1.SignInResp{
			AccessToken:  accessToken.Token,
			RefreshToken: refreshToken.Token,
			UserInfo:     user,
		}
		return nil
	})
	return resp, err
}

func (u *UserService) SignOut(ctx context.Context, req *userv1.SignOutReq) (*userv1.SignOutResp, error) {
	claims, err := u.token.Parse(req.AccessToken, enumsv1.TokenType_TOKEN_TYPE_ACCESS.String())
	if err != nil {
		return nil, err
	}
	accessJti := common.GetJwtJti(claims)
	if accessJti == "" {
		return nil, ecode.ErrUserTokenInvalid
	}
	logQ := u.db.UserSignLog
	logModel, err := logQ.WithContext(ctx).Where(
		logQ.AccessJti.Eq(accessJti),
		logQ.Status.Neq(int16(enumsv1.SignInStatus_SIGN_IN_STATUS_OK)),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrUserTokenInvalid
		}
		return nil, err
	}
	if _, err := logQ.WithContext(ctx).Where(logQ.AccessJti.Eq(accessJti)).Update(logQ.Status, int16(enumsv1.SignInStatus_SIGN_IN_STATUS_SIGN_OUT)); err != nil {
		return nil, err
	}
	if err := u.addJtiToBlkByLog(ctx, logModel); err != nil {
		return nil, err
	}
	return &userv1.SignOutResp{}, nil
}

func (u *UserService) ValidateToken(ctx context.Context, req *userv1.ValidateTokenReq) (*userv1.ValidateTokenResp, error) {
	claims, err := u.token.Parse(req.Token, req.Type.String())
	if err != nil {
		return nil, err
	}
	jwtClaims := common.ClaimsToJwtClaims(claims, req.ExtraKey)
	blocked, err := u.blk.Exists(ctx, jwtClaims.Jti)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ecode.ErrUserTokenInvalid
	}
	return &userv1.ValidateTokenResp{
		Type:   req.Type,
		Claims: jwtClaims,
	}, nil
}

func (u *UserService) RefreshToken(ctx context.Context, req *userv1.RefreshTokenReq) (*userv1.RefreshTokenResp, error) {
	claims, err := u.token.Parse(req.RefreshToken, enumsv1.TokenType_TOKEN_TYPE_REFRESH.String())
	if err != nil {
		return nil, err
	}
	jti := common.GetJwtJti(claims)
	if jti == "" {
		return nil, ecode.ErrUserTokenInvalid
	}
	blocked, err := u.blk.Exists(ctx, jti)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ecode.ErrUserTokenInvalid
	}
	sub, err := claims.GetSubject()
	if err != nil {
		return nil, err
	}
	uid, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return nil, err
	}
	accountQ := u.db.UserAccount
	userAccount, err := accountQ.WithContext(ctx).Where(accountQ.ID.Eq(uid)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrUserNotFound
		}
		return nil, err
	}
	if !common.IsUserValid(userAccount.Status) {
		return nil, ecode.ErrUserDisabled
	}
	user := common.UserModelToUser(userAccount)
	accessToken, refreshToken, err := u.genToken(user, req.ExtraJwtClaims)
	if err != nil {
		return nil, err
	}
	logQ := u.db.UserSignLog
	if _, err := logQ.WithContext(ctx).Where(logQ.RefreshJti.Eq(jti)).Updates(
		&model.UserSignLog{
			AccessJti:        accessToken.Jti,
			RefreshJti:       refreshToken.Jti,
			AccessExpiredAt:  trans.Ref(accessToken.Exp),
			RefreshExpiredAt: trans.Ref(refreshToken.Exp),
		},
	); err != nil {
		return nil, err
	}
	if err := u.addJtiToBlkByClaims(ctx, claims); err != nil {
		return nil, err
	}
	return &userv1.RefreshTokenResp{
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken.Token,
		UserInfo:     user,
	}, nil
}

func (u *UserService) addJtiToBlkByLog(ctx context.Context, logModel *model.UserSignLog) error {
	var items []*blocklist.BlockItem
	now := time.Now()
	if logModel.AccessExpiredAt != nil {
		ttl := logModel.AccessExpiredAt.Sub(now)
		if ttl > 0 {
			items = append(items, &blocklist.BlockItem{
				Target: logModel.AccessJti,
				TTL:    ttl,
			})
		}
	}
	if logModel.RefreshExpiredAt != nil {
		ttl := logModel.RefreshExpiredAt.Sub(now)
		if ttl > 0 {
			items = append(items, &blocklist.BlockItem{
				Target: logModel.RefreshJti,
				TTL:    ttl,
			})
		}
	}
	return u.blk.BatchAdd(ctx, items)
}

func (u *UserService) addJtiToBlkByClaims(ctx context.Context, claims jwtv5.MapClaims) error {
	exp := common.GetJwtExp(claims)
	if exp == nil {
		return ecode.ErrUserTokenInvalid
	}
	ttl := exp.Sub(time.Now())
	if ttl <= 0 {
		return nil
	}
	jti := common.GetJwtJti(claims)
	return u.blk.Add(ctx, jti, ttl)
}

func (u *UserService) genToken(user *userv1.User, extra map[string]string) (accessToken, refreshToken *jwt.Token, err error) {
	extraClaims := map[string]any{
		common.JwtTenantIDKey: user.GetTenantId(),
		common.JwtNumberKey:   user.GetNumber(),
		common.JwtTypeKey:     user.GetUserType(),
		common.JwtLevelKey:    user.GetUserLevel(),
	}
	for k, v := range extra {
		extraClaims[k] = v
	}
	if accessToken, err = u.token.Generate(
		strconv.FormatInt(user.Uid, 10),
		enumsv1.TokenType_TOKEN_TYPE_ACCESS.String(),
		u.cfg.Jwt.AccessTtl.AsDuration(),
		extraClaims,
	); err != nil {
		return
	}
	if refreshToken, err = u.token.Generate(
		strconv.FormatInt(user.Uid, 10),
		enumsv1.TokenType_TOKEN_TYPE_REFRESH.String(),
		u.cfg.Jwt.RefreshTtl.AsDuration(),
		extraClaims,
	); err != nil {
		return
	}
	return
}

func (u *UserService) getAuthProvider(typ enumsv1.SignInType) (auth.Auth, error) {
	provider, ok := u.authProviders[typ]
	if !ok {
		return nil, fmt.Errorf("unknown sign-in type %s", typ)
	}
	return provider, nil
}

func newAuthProvider(config *configv1.Config) map[enumsv1.SignInType]auth.Auth {
	authMap := make(map[enumsv1.SignInType]auth.Auth, len(config.User.Auth))
	for _, v := range config.User.Auth {
		switch v.Type {
		case enumsv1.SignInType_SIGN_IN_TYPE_WECHAT_MINI:
			authMap[v.Type] = tencent.NewWechatManager(v.Wechat)
		case enumsv1.SignInType_SIGN_IN_TYPE_HUAWEI:
			shortID := common.NewIDService(global_id.NewOnce(config), singleton.NewShortID(config.ShortId))
			authMap[v.Type] = huawei.NewAccountManager(
				config.User.KeyPrefix,
				singleton.NewRDB(config.Redis),
				shortID,
				v.Huawei,
			)
		}
	}
	return authMap
}
