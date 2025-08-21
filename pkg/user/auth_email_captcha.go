package user

import (
	"context"

	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	messagev1 "github.com/byteflowing/base/gen/message/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/base/pkg/captcha"
	"github.com/byteflowing/go-common/trans"
)

type EmailCaptcha struct {
	captcha    captcha.Captcha
	repo       Repo
	jwtService *JwtService
}

func NewEmailCaptcha(captcha captcha.Captcha, repo Repo, jwtService *JwtService) Authenticator {
	return &EmailCaptcha{
		captcha:    captcha,
		repo:       repo,
		jwtService: jwtService,
	}
}

func (e *EmailCaptcha) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_EMAIL_CAPTCHA
}

func (e *EmailCaptcha) Authenticate(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != e.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.CaptchaToken == nil {
		return nil, ecode.ErrUserCaptchaTokenIsEmpty
	}
	if len(req.Credential) == 0 {
		return nil, ecode.ErrUserCaptchaIsEmpty
	}
	_, err = e.captcha.Verify(ctx, &messagev1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL,
		Token:      trans.Deref(req.CaptchaToken),
		Captcha:    req.Credential,
		Number:     &messagev1.VerifyCaptchaReq_Email{Email: req.Identifier},
	})
	if err != nil {
		return nil, err
	}
	userBasic, err := e.repo.GetUserBasicByEmail(ctx, req.Identifier)
	if err != nil {
		return nil, err
	}
	if isDisabled(userBasic) {
		return nil, ecode.ErrUserDisabled
	}
	// 生成jwt token
	accessToken, refreshToken, err := e.jwtService.GenerateToken(ctx, &GenerateJwtReq{
		UserBasic:      userBasic,
		SignInReq:      req,
		ExtraJwtClaims: req.ExtraJwtClaims,
		AuthType:       e.AuthType(),
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
