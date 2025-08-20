package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/base/pkg/common"
)

type WeChat struct {
	repo       Repo
	jwtService *JwtService
	wechatMgr  *common.WechatManager
}

func NewWeChat(repo Repo, jwtService *JwtService, wechatMgr *common.WechatManager) Authenticator {
	return &WeChat{
		repo:       repo,
		jwtService: jwtService,
		wechatMgr:  wechatMgr,
	}
}

func (w *WeChat) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_WECHAT
}

func (w *WeChat) Authenticate(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != enumsv1.AuthType_AUTH_TYPE_WECHAT {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	res, err := w.wechatMgr.WechatSignIn(ctx, &commonv1.WechatSignInReq{
		Appid: req.Identifier,
		Code:  req.Credential,
	})
	if err != nil {
		return nil, err
	}
	userBasic, err := w.repo.GetOrCreateUserAuthByWechat(ctx, res.Appid, res.Openid, res.UnionId, res.SessionKey)
	if err != nil {
		return nil, err
	}
	if isDisabled(userBasic) {
		return nil, ecode.ErrUserDisabled
	}
	// 生成jwt token
	accessToken, refreshToken, err := w.jwtService.GenerateToken(ctx, &GenerateJwtReq{
		UserBasic:      userBasic,
		SignInReq:      req,
		ExtraJwtClaims: req.ExtraJwtClaims,
		AuthType:       enumsv1.AuthType_AUTH_TYPE_WECHAT,
		AppId:          res.Appid,
		OpenId:         res.Openid,
		UnionId:        res.UnionId,
	})
	if err != nil {
		return nil, err
	}
	resp = &userv1.SignInResp{
		Data: &userv1.SignInResp_Data{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}
	return resp, nil
}
