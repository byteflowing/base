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

type PhoneCaptcha struct {
	captcha    captcha.Captcha
	repo       Repo
	jwtService *JwtService
}

func NewPhoneCaptcha(captcha captcha.Captcha, repo Repo, jwtService *JwtService) Authenticator {
	return &PhoneCaptcha{
		captcha:    captcha,
		repo:       repo,
		jwtService: jwtService,
	}
}

func (p *PhoneCaptcha) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_PHONE_CAPTCHA
}

func (p *PhoneCaptcha) Authenticate(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != p.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.CaptchaToken == nil {
		return nil, ecode.ErrUserCaptchaTokenIsEmpty
	}
	if len(req.Credential) == 0 {
		return nil, ecode.ErrUserCaptchaIsEmpty
	}
	if req.PhoneNumber == nil || req.PhoneNumber.Number == "" || req.PhoneNumber.CountryCode == "" {
		return nil, ecode.ErrPhoneIsEmpty
	}
	_, err = p.captcha.Verify(ctx, &messagev1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS,
		Token:      trans.Deref(req.CaptchaToken),
		Captcha:    req.Credential,
		Number:     &messagev1.VerifyCaptchaReq_PhoneNumber{PhoneNumber: req.PhoneNumber},
	})
	if err != nil {
		return nil, err
	}
	userBasic, err := p.repo.GetUserBasicByPhone(ctx, req.PhoneNumber)
	if err != nil {
		return nil, err
	}
	if isDisabled(userBasic) {
		return nil, ecode.ErrUserDisabled
	}
	// 生成jwt token
	accessToken, refreshToken, err := p.jwtService.GenerateToken(ctx, &GenerateJwtReq{
		UserBasic:      userBasic,
		SignInReq:      req,
		ExtraJwtClaims: req.ExtraJwtClaims,
		AuthType:       p.AuthType(),
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
