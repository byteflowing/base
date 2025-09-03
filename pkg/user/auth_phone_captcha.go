package user

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/byteflowing/base/dal/query"
	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	messagev1 "github.com/byteflowing/base/gen/msg/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/base/pkg/captcha"
	"github.com/byteflowing/base/pkg/common"
	"github.com/byteflowing/go-common/crypto"
	"github.com/byteflowing/go-common/trans"
)

type PhoneCaptcha struct {
	captcha        captcha.Captcha
	repo           Repo
	jwtService     *JwtService
	shortIDGen     *common.ShortIDGenerator
	globalIDGen    common.GlobalIdGenerator
	passwordHasher *crypto.PasswordHasher
}

func NewPhoneCaptcha(
	captcha captcha.Captcha,
	repo Repo,
	jwtService *JwtService,
	shortIDGen *common.ShortIDGenerator,
	globalIDGen common.GlobalIdGenerator,
	passwordHasher *crypto.PasswordHasher,
) Authenticator {
	return &PhoneCaptcha{
		captcha:        captcha,
		repo:           repo,
		jwtService:     jwtService,
		shortIDGen:     shortIDGen,
		globalIDGen:    globalIDGen,
		passwordHasher: passwordHasher,
	}
}

func (p *PhoneCaptcha) AuthType() enumsv1.AuthType {
	return enumsv1.AuthType_AUTH_TYPE_PHONE_CAPTCHA
}

func (p *PhoneCaptcha) SignUp(ctx context.Context, tx *query.Query, req *userv1.SignUpReq) (resp *userv1.SignUpResp, err error) {
	if req.AuthType != p.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.CaptchaToken == "" {
		return nil, ecode.ErrUserCaptchaTokenIsEmpty
	}
	_, err = p.captcha.VerifyCaptcha(ctx, &messagev1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS,
		Token:      req.CaptchaToken,
		Captcha:    req.Captcha,
		Number:     &messagev1.VerifyCaptchaReq_PhoneNumber{PhoneNumber: req.Phone},
	})
	if err != nil {
		return nil, err
	}
	if err = checkUserBasicUnique(ctx, tx, p.repo, req.Biz, req.Phone, req.Number, req.Email); err != nil {
		return nil, err
	}
	userBasic, err := signUpReqToUserBasic(req, p.globalIDGen, p.shortIDGen, p.passwordHasher)
	if err != nil {
		return nil, err
	}
	if err = p.repo.CreateUserBasic(ctx, tx, userBasic); err != nil {
		return nil, err
	}
	resp = &userv1.SignUpResp{
		Data: &userv1.SignUpResp_Data{UserInfo: userBasicToUserInfo(userBasic)},
	}
	return resp, nil
}

func (p *PhoneCaptcha) SignIn(ctx context.Context, tx *query.Query, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
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
	_, err = p.captcha.VerifyCaptcha(ctx, &messagev1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS,
		Token:      trans.Deref(req.CaptchaToken),
		Captcha:    req.Credential,
		Number:     &messagev1.VerifyCaptchaReq_PhoneNumber{PhoneNumber: req.PhoneNumber},
	})
	if err != nil {
		return nil, err
	}
	userBasic, err := p.repo.GetUserBasicByPhone(ctx, tx, req.Biz, req.PhoneNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrPhoneNotExist
		}
		return nil, err
	}
	err = tx.Transaction(func(tx *query.Query) error {
		resp, err = checkPasswordAndGenToken(ctx, tx, req, userBasic, p.jwtService, nil, nil)
		if err != nil {
			return err
		}
		verified := int16(enumsv1.Verified_VERIFIED_VERIFIED)
		if userBasic.PhoneVerified != verified {
			userBasic.PhoneVerified = verified
			return p.repo.UpdateUserBasicByUid(ctx, tx, userBasic)
		}
		return nil
	})
	return
}

func (p *PhoneCaptcha) SignOut(ctx context.Context, tx *query.Query, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error) {
	if req.AuthType != p.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	return signOutBySessionId(ctx, req, p.repo, tx, p.jwtService)
}

func (p *PhoneCaptcha) Bind(ctx context.Context, tx *query.Query, req *userv1.BindUserAuthReq) (resp *userv1.BindUserAuthResp, err error) {
	if req.Type != p.AuthType() {
		return nil, ecode.ErrUserAuthTypeMisMatch
	}
	if req.CaptchaToken == nil {
		return nil, ecode.ErrUserCaptchaTokenIsEmpty
	}
	if req.Captcha == nil || len(*req.Captcha) == 0 {
		return nil, ecode.ErrUserCaptchaIsEmpty
	}
	if req.Phone == nil || req.Phone.Number == "" || req.Phone.CountryCode == "" {
		return nil, ecode.ErrPhoneIsEmpty
	}
	_, err = p.captcha.VerifyCaptcha(ctx, &messagev1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS,
		Token:      trans.Deref(req.CaptchaToken),
		Captcha:    trans.StringValue(req.Captcha),
		Number:     &messagev1.VerifyCaptchaReq_PhoneNumber{PhoneNumber: req.Phone},
	})
	if err != nil {
		return nil, err
	}
	exist, err := p.repo.CheckPhoneExists(ctx, tx, req.Biz, req.Phone)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, ecode.ErrPhoneExists
	}
	userBasic, err := p.repo.GetUserBasicByUID(ctx, tx, req.Uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ecode.ErrUserNotExist
		}
		return nil, err
	}
	if len(userBasic.Phone) == 0 {
		return nil, ecode.ErrPhoneAlreadyBind
	}
	userBasic.PhoneCountryCode = req.Phone.CountryCode
	userBasic.Phone = req.Phone.Number
	userBasic.PhoneVerified = int16(enumsv1.Verified_VERIFIED_VERIFIED)
	if err = p.repo.UpdateUserBasicByUid(ctx, tx, userBasic); err != nil {
		return nil, err
	}
	return nil, nil
}
