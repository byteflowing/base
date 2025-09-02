package user

import (
	"context"
	"errors"

	"github.com/byteflowing/base/dal/query"
	"gorm.io/gorm"

	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	messagev1 "github.com/byteflowing/base/gen/msg/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/base/pkg/captcha"
	"github.com/byteflowing/base/pkg/common"
	"github.com/byteflowing/go-common/crypto"
	"github.com/byteflowing/go-common/trans"
)

type EmailCaptcha struct {
	captcha        captcha.Captcha
	repo           Repo
	jwtService     *JwtService
	shortIDGen     *common.ShortIDGenerator
	globalIDGen    common.GlobalIdGenerator
	passwordHasher *crypto.PasswordHasher
}

func NewEmailCaptcha(
	captcha captcha.Captcha,
	repo Repo,
	jwtService *JwtService,
	shortIDGen *common.ShortIDGenerator,
	globalIDGen common.GlobalIdGenerator,
	passwordHasher *crypto.PasswordHasher,
) Authenticator {
	return &EmailCaptcha{
		captcha:        captcha,
		repo:           repo,
		jwtService:     jwtService,
		shortIDGen:     shortIDGen,
		globalIDGen:    globalIDGen,
		passwordHasher: passwordHasher,
	}
}

func (e *EmailCaptcha) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_EMAIL_CAPTCHA
}

func (e *EmailCaptcha) SignUp(ctx context.Context, tx *query.Query, req *userv1.SignUpReq) (resp *userv1.SignUpResp, err error) {
	if req.AuthType != e.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.CaptchaToken == "" {
		return nil, ecode.ErrUserCaptchaTokenIsEmpty
	}
	if _, err = e.captcha.VerifyCaptcha(ctx, &messagev1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL,
		Token:      req.CaptchaToken,
		Captcha:    req.Captcha,
		Number:     &messagev1.VerifyCaptchaReq_Email{Email: trans.StringValue(req.Email)},
	}); err != nil {
		return nil, err
	}
	if err = checkUserBasicUnique(ctx, req, e.repo, tx); err != nil {
		return nil, err
	}
	userBasic, err := signUpReqToUserBasic(req, e.globalIDGen, e.shortIDGen, e.passwordHasher)
	if err != nil {
		return nil, err
	}
	if err = e.repo.CreateUserBasic(ctx, tx, userBasic); err != nil {
		return nil, err
	}
	resp = &userv1.SignUpResp{
		Data: &userv1.SignUpResp_Data{UserInfo: userBasicToUserInfo(userBasic)},
	}
	return resp, nil
}

func (e *EmailCaptcha) SignIn(ctx context.Context, tx *query.Query, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	if req.AuthType != e.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.CaptchaToken == nil {
		return nil, ecode.ErrUserCaptchaTokenIsEmpty
	}
	if len(req.Credential) == 0 {
		return nil, ecode.ErrUserCaptchaIsEmpty
	}
	if _, err = e.captcha.VerifyCaptcha(ctx, &messagev1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL,
		Token:      trans.Deref(req.CaptchaToken),
		Captcha:    req.Credential,
		Number:     &messagev1.VerifyCaptchaReq_Email{Email: req.Identifier},
	}); err != nil {
		return nil, err
	}
	userBasic, err := e.repo.GetUserBasicByEmail(ctx, tx, req.Biz, req.Identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrEmailNotExist
		}
		return nil, err
	}
	err = tx.Transaction(func(tx *query.Query) error {
		resp, err = checkPasswordAndGenToken(ctx, tx, req, userBasic, e.jwtService, nil, nil)
		if err != nil {
			return err
		}
		if userBasic.EmailVerified != int16(enumsv1.Verified_VERIFIED_VERIFIED) {
			userBasic.EmailVerified = int16(enumsv1.Verified_VERIFIED_VERIFIED)
			return e.repo.UpdateUserBasicByUid(ctx, tx, userBasic)
		}
		return nil
	})
	return
}

func (e *EmailCaptcha) SignOut(ctx context.Context, tx *query.Query, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error) {
	if req.AuthType != e.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	return signOutBySessionId(ctx, req, e.repo, tx, e.jwtService)
}
