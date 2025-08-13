package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/message/mail"
)

type EmailCaptcha struct {
	mail       mail.Mail
	repo       Repo
	jwtService *JwtService
}

func NewEmailCaptcha(mail mail.Mail, repo Repo, jwtService *JwtService) Authenticator {
	return &EmailCaptcha{
		mail:       mail,
		repo:       repo,
		jwtService: jwtService,
	}
}

func (e *EmailCaptcha) AuthType() AuthType {
	return AuthTypeEmailCaptcha
}

func (e *EmailCaptcha) Authenticate(ctx context.Context, req *SignInReq) (resp *SignInResp, err error) {
	if req.AuthType != AuthTypeEmailCaptcha {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.CaptchaToken == "" {
		return nil, ecode.ErrUserCaptchaTokenIsEmpty
	}
	if len(req.Credential) == 0 {
		return nil, ecode.ErrUserCaptchaIsEmpty
	}
	ok, err := e.mail.VerifyCaptcha(ctx, &mail.VerifyCaptchaReq{
		Token:   req.CaptchaToken,
		Captcha: req.Credential,
	})
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ecode.ErrUserCaptchaMisMatch
	}
	userBasic, err := e.repo.GetUserBasicByEmail(ctx, req.Identifier)
	if err != nil {
		return nil, err
	}
	// 生成jwt token
	accessToken, refreshToken, err := e.jwtService.GenerateToken(ctx, userBasic, req)
	if err != nil {
		return nil, err
	}
	resp = &SignInResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return resp, nil
}
