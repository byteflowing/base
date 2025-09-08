package user

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/byteflowing/base/dal/query"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/captcha"
	"github.com/byteflowing/base/pkg/common"
	"github.com/byteflowing/go-common/crypto"
	"github.com/byteflowing/go-common/trans"
	captchav1 "github.com/byteflowing/proto/gen/go/captcha/v1"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	userv1 "github.com/byteflowing/proto/gen/go/services/user/v1"
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
	if _, err = e.captcha.VerifyCaptcha(ctx, &captchav1.VerifyCaptchaReq{
		SenderType:  enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL,
		Token:       req.CaptchaToken,
		Captcha:     req.Captcha,
		CaptchaType: req.CaptchaType,
		Number:      &captchav1.VerifyCaptchaReq_Email{Email: trans.StringValue(req.Email)},
	}); err != nil {
		return nil, err
	}
	if err = checkUserBasicUnique(ctx, tx, e.repo, req.Biz, req.Phone, req.Number, req.Email); err != nil {
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
	if _, err = e.captcha.VerifyCaptcha(ctx, &captchav1.VerifyCaptchaReq{
		SenderType:  enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL,
		Token:       trans.Deref(req.CaptchaToken),
		Captcha:     req.Credential,
		CaptchaType: req.CaptchaType,
		Number:      &captchav1.VerifyCaptchaReq_Email{Email: req.Identifier},
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
		verified := int16(enumsv1.Verified_VERIFIED_VERIFIED)
		if userBasic.EmailVerified != verified {
			userBasic.EmailVerified = verified
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

func (e *EmailCaptcha) Bind(ctx context.Context, tx *query.Query, req *userv1.BindUserAuthReq) (resp *userv1.BindUserAuthResp, err error) {
	if req.Type != e.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.CaptchaToken == nil {
		return nil, ecode.ErrUserCaptchaTokenIsEmpty
	}
	if req.Email == nil {
		return nil, ecode.ErrEmailIsEmpty
	}
	if _, err = e.captcha.VerifyCaptcha(ctx, &captchav1.VerifyCaptchaReq{
		SenderType:  enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL,
		Token:       trans.Deref(req.CaptchaToken),
		Captcha:     trans.StringValue(req.Captcha),
		CaptchaType: trans.StringValue(req.CaptchaType),
		Number:      &captchav1.VerifyCaptchaReq_Email{Email: trans.StringValue(req.Email)},
	}); err != nil {
		return nil, err
	}
	exist, err := e.repo.CheckEmailExists(ctx, tx, req.Biz, *req.Email)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, ecode.ErrEmailExists
	}
	userBasic, err := e.repo.GetUserBasicByUID(ctx, tx, req.Uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrUserNotExist
		}
		return nil, err
	}
	if len(userBasic.Email) > 0 {
		return nil, ecode.ErrEmailAlreadyBind
	}
	userBasic.Email = trans.Deref(req.Email)
	userBasic.EmailVerified = int16(enumsv1.Verified_VERIFIED_VERIFIED)
	if err = e.repo.UpdateUserBasicByUid(ctx, tx, userBasic); err != nil {
		return nil, err
	}
	return nil, nil
}
