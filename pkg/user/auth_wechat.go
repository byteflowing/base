package user

import (
	"context"

	"github.com/byteflowing/base/dal/model"
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

func (w *WeChat) Authenticate(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
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
	userAuth, err = w.repo.GetUserAuthByOpenID(ctx, res.Openid)
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
		if err = w.repo.UpdateUserAuth(ctx, userAuth); err != nil {
			return nil, err
		}
		if userBasic, err = w.repo.GetUserBasicByUID(ctx, userAuth.UID); err != nil {
			return nil, err
		}
	} else {
		// 4. 如果没有找到，再通过unionid去查找
		if res.UnionId != "" {
			userAuth, err = w.repo.GetUserAuthByUnionID(ctx, res.UnionId)
			if err != nil {
				return nil, err
			}
			// 5. 如果找到userAuth说明是老用户,只是第一次使用appid登录
			if userAuth != nil {
				userBasic, err = w.repo.GetUserBasicByUID(ctx, userAuth.UID)
				userAuth = &model.UserAuth{
					UID:        userAuth.UID,
					Type:       int16(enumsv1.AuthType_AUTH_TYPE_WECHAT),
					Status:     int16(enumsv1.AuthStatus_AUTH_STATUS_OK),
					Appid:      res.Appid,
					Identifier: res.Openid,
					Credential: res.SessionKey,
					UnionID:    trans.Ref(res.UnionId),
				}
				if err = w.repo.CreateUserAuth(ctx, userAuth); err != nil {
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
				ID:     uid,
				Number: number,
				Status: int16(enumsv1.UserStatus_USER_STATUS_OK),
				Source: int16(enumsv1.UserSource_USER_SOURCE_WECHAT),
			}
			userAuth = &model.UserAuth{
				Type:       int16(enumsv1.AuthType_AUTH_TYPE_WECHAT),
				Status:     int16(enumsv1.AuthStatus_AUTH_STATUS_OK),
				Appid:      res.Appid,
				Identifier: res.Openid,
				Credential: res.SessionKey,
				UnionID:    trans.Ref(res.UnionId),
			}
			if err = w.repo.CreateUserBasicAndAuth(ctx, userBasic, userAuth); err != nil {
				return nil, err
			}
		}
	}
	return checkPasswordAndGenToken(ctx, req, userBasic, w.jwtService, nil, nil)
}
