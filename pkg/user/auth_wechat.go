package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/go-common/3rd/tencent/mini"
)

type WeChat struct {
	repo       Repo
	jwtService *JwtService
	mini       *mini.Client
}

func NewWeChat(repo Repo, jwtService *JwtService, mini *mini.Client) Authenticator {
	return &WeChat{
		repo:       repo,
		jwtService: jwtService,
		mini:       mini,
	}
}

func (w *WeChat) AuthType() AuthType {
	return AuthTypeWechat
}

func (w *WeChat) Authenticate(ctx context.Context, req *SignInReq) (resp *SignInResp, err error) {
	if req.AuthType != AuthTypeWechat {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	res, err := w.mini.WechatLogin(ctx, &mini.WechatLoginReq{Code: req.Identifier})
	if err != nil {
		return nil, err
	}
	userBasic, err := w.repo.GetOrCreateUserAuthByWechat(ctx, res.OpenID, res.UnionID, res.SessionKey)
	if err != nil {
		return nil, err
	}
	// 生成jwt token
	accessToken, refreshToken, err := w.jwtService.GenerateToken(ctx, userBasic, req)
	if err != nil {
		return nil, err
	}
	resp = &SignInResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return resp, nil
}
