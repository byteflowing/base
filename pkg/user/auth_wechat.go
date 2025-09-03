package user

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/dal/query"
	"github.com/byteflowing/base/ecode"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/base/pkg/common"
	"github.com/byteflowing/go-common/trans"
)

type WeChat struct {
	repo        Repo
	jwtService  *JwtService
	wechatMgr   *common.WechatManager
	shortIDGen  *common.ShortIDGenerator
	globalIDGen common.GlobalIdGenerator
}

func NewWeChat(
	repo Repo,
	jwtService *JwtService,
	wechatMgr *common.WechatManager,
	shortIDGen *common.ShortIDGenerator,
	globalIDGen common.GlobalIdGenerator,
) Authenticator {
	return &WeChat{
		repo:        repo,
		jwtService:  jwtService,
		wechatMgr:   wechatMgr,
		shortIDGen:  shortIDGen,
		globalIDGen: globalIDGen,
	}
}

func (w *WeChat) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_WECHAT
}

func (w *WeChat) SignUp(ctx context.Context, tx *query.Query, req *userv1.SignUpReq) (*userv1.SignUpResp, error) {
	return nil, ecode.ErrUnImplemented
}

func (w *WeChat) SignIn(ctx context.Context, tx *query.Query, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != w.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	// 1. 换取微信登录信息
	res, err := w.wechatMgr.WechatSignIn(ctx, &commonv1.WechatSignInReq{
		Appid: req.Identifier,
		Code:  req.Credential,
	})
	if err != nil {
		return nil, err
	}
	var (
		userAuth  *model.UserAuth
		userBasic *model.UserBasic
	)

	// 2. 通过openid查找用户
	userAuth, err = w.repo.GetUserAuthByOpenID(ctx, tx, res.Openid)
	if err != nil {
		return nil, err
	}
	// 3. 如果找到userAuth说明用户存在，直接更新
	if userAuth != nil {
		if isAuthDisabled(userAuth) {
			return nil, ecode.ErrUserAuthDisabled
		}
		userAuth.Credential = res.SessionKey
		userAuth.UnionID = trans.Ref(res.UnionId)
		if err = w.repo.UpdateUserAuth(ctx, tx, userAuth); err != nil {
			return nil, err
		}
		if userBasic, err = w.repo.GetUserBasicByUID(ctx, tx, userAuth.UID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ecode.ErrUserNotExist
			}
			return nil, err
		}
	} else {
		// 4. 如果没有找到，再通过unionid去查找
		if res.UnionId != "" {
			userAuth, err = w.repo.GetOneUserAuthByUnionID(ctx, tx, req.Biz, res.UnionId)
			if err != nil {
				return nil, err
			}
			// 5. 如果找到userAuth说明是老用户,只是第一次使用appid登录
			if userAuth != nil {
				userBasic, err = w.repo.GetUserBasicByUID(ctx, tx, userAuth.UID)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return nil, ecode.ErrUserNotExist
					}
					return nil, err
				}
				userAuth = &model.UserAuth{
					Biz:        req.Biz,
					UID:        userAuth.UID,
					Type:       int16(enumsv1.AuthType_AUTH_TYPE_WECHAT),
					Status:     int16(enumsv1.AuthStatus_AUTH_STATUS_OK),
					Appid:      res.Appid,
					Identifier: res.Openid,
					Credential: res.SessionKey,
					UnionID:    trans.Ref(res.UnionId),
				}
				if err = w.repo.CreateUserAuth(ctx, tx, userAuth); err != nil {
					return nil, err
				}
			}
		}
		// 6. 如果unionid也没有找到说明是一个全新用户
		if userAuth == nil {
			uid, err := w.globalIDGen.GetID()
			if err != nil {
				return nil, err
			}
			number, err := w.shortIDGen.GetID()
			if err != nil {
				return nil, err
			}
			userBasic = &model.UserBasic{
				ID:             uid,
				Number:         number,
				Biz:            req.Biz,
				Status:         int16(enumsv1.UserStatus_USER_STATUS_OK),
				Source:         int16(enumsv1.UserSource_USER_SOURCE_WECHAT),
				RegisterIP:     req.Ip,
				RegisterDevice: req.Device,
				RegisterAgent:  req.UserAgent,
			}
			if err = tx.Transaction(func(tx *query.Query) error {
				if err := tx.UserBasic.WithContext(ctx).Create(userBasic); err != nil {
					return err
				}
				userAuth = &model.UserAuth{
					Biz:        req.Biz,
					UID:        userBasic.ID,
					Type:       int16(enumsv1.AuthType_AUTH_TYPE_WECHAT),
					Status:     int16(enumsv1.AuthStatus_AUTH_STATUS_OK),
					Appid:      res.Appid,
					Identifier: res.Openid,
					Credential: res.SessionKey,
					UnionID:    trans.Ref(res.UnionId),
				}
				return tx.UserAuth.WithContext(ctx).Create(userAuth)
			}); err != nil {
				return nil, err
			}
		}
	}
	return checkPasswordAndGenToken(ctx, tx, req, userBasic, w.jwtService, nil, nil)
}

func (w *WeChat) SignOut(ctx context.Context, tx *query.Query, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error) {
	if req.AuthType != w.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	return signOutBySessionId(ctx, req, w.repo, tx, w.jwtService)
}

func (w *WeChat) Bind(ctx context.Context, tx *query.Query, req *userv1.BindUserAuthReq) (resp *userv1.BindUserAuthResp, err error) {
	if req.Type != w.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	// 1. 换取微信登录信息
	res, err := w.wechatMgr.WechatSignIn(ctx, &commonv1.WechatSignInReq{
		Appid: trans.StringValue(req.AppId),
		Code:  trans.StringValue(req.Code),
	})
	if err != nil {
		return nil, err
	}
	userAuth, err := w.repo.GetUserAuthByOpenID(ctx, tx, res.Openid)
	if err != nil {
		return nil, err
	}
	if userAuth == nil {
		auth := &model.UserAuth{
			UID:        req.Uid,
			Type:       int16(enumsv1.AuthType_AUTH_TYPE_WECHAT),
			Status:     int16(enumsv1.AuthStatus_AUTH_STATUS_OK),
			Appid:      trans.StringValue(req.AppId),
			Biz:        req.Biz,
			Identifier: res.Openid,
			Credential: res.SessionKey,
		}
		if res.UnionId != "" {
			auth.UnionID = trans.String(res.UnionId)
		}
		if err := w.repo.CreateUserAuth(ctx, tx, auth); err != nil {
			return nil, err
		}
		return nil, nil
	}
	userAuth.Credential = res.SessionKey
	if res.UnionId != "" {
		userAuth.UnionID = trans.String(res.UnionId)
	}
	if err = w.repo.UpdateUserAuth(ctx, tx, userAuth); err != nil {
		return nil, err
	}
	return nil, nil
}
