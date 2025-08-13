package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/message/sms"
)

type PhoneCaptcha struct {
	sms        sms.Sms
	repo       Repo
	jwtService *JwtService
}

func NewPhoneCaptcha(sms sms.Sms, repo Repo, jwtService *JwtService) Authenticator {
	return &PhoneCaptcha{
		sms:        sms,
		repo:       repo,
		jwtService: jwtService,
	}
}

func (p *PhoneCaptcha) AuthType() AuthType {
	return AuthTypePhoneCaptcha
}

func (p *PhoneCaptcha) Authenticate(ctx context.Context, req *SignInReq) (resp *SignInResp, err error) {
	if req.AuthType != AuthTypePhoneCaptcha {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.CaptchaToken == "" {
		return nil, ecode.ErrUserCaptchaTokenIsEmpty
	}
	if len(req.Credential) == 0 {
		return nil, ecode.ErrUserCaptchaIsEmpty
	}
	ok, err := p.sms.VerifyCaptcha(ctx, &sms.VerifyCaptchaReq{
		Token:   req.CaptchaToken,
		Captcha: req.Credential,
	})
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ecode.ErrUserCaptchaMisMatch
	}
	userBasic, err := p.repo.GetUserBasicByPhone(ctx, req.Identifier)
	if err != nil {
		return nil, err
	}
	// 生成jwt token
	accessToken, refreshToken, err := p.jwtService.GenerateToken(ctx, userBasic, req)
	if err != nil {
		return nil, err
	}
	resp = &SignInResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return resp, nil
}
