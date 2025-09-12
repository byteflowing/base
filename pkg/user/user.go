package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/dal/query"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/captcha"
	"github.com/byteflowing/base/pkg/common"
	"github.com/byteflowing/go-common/config"
	"github.com/byteflowing/go-common/crypto"
	"github.com/byteflowing/go-common/idx"
	"github.com/byteflowing/go-common/trans"
	captchav1 "github.com/byteflowing/proto/gen/go/captcha/v1"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	userv1 "github.com/byteflowing/proto/gen/go/services/user/v1"
)

type Authenticator interface {
	AuthType() enumsv1.AuthType
	SignUp(ctx context.Context, tx *query.Query, req *userv1.SignUpReq) (*userv1.SignUpResp, error)
	SignIn(ctx context.Context, tx *query.Query, req *userv1.SignInReq) (resp *userv1.SignInResp, err error)
	SignOut(ctx context.Context, tx *query.Query, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error)
	Bind(ctx context.Context, tx *query.Query, req *userv1.BindUserAuthReq) (resp *userv1.BindUserAuthResp, err error)
}

type Impl struct {
	config        *configv1.Config
	authHandlers  map[enumsv1.AuthType]Authenticator
	repo          Repo
	query         *query.Query
	passHasher    *crypto.PasswordHasher
	jwtService    *JwtService
	tokenVerifier *TwoStepVerifier
	captcha       captcha.Captcha
	shortIDGen    *common.ShortIDGenerator
	globalIDGen   common.GlobalIdGenerator
	grpServer     *grpc.Server
	userv1.UnimplementedUserServiceServer
}

func NewConfig(filePath string) *configv1.Config {
	conf := &configv1.Config{}
	if err := config.ReadProtoConfig(filePath, conf); err != nil {
		panic(err)
	}
	return conf
}

func New(
	config *configv1.Config,
	repo Repo,
	query *query.Query,
	jwtService *JwtService,
	tokenVerifier *TwoStepVerifier,
	shortIDGen *common.ShortIDGenerator,
	globalIDGen common.GlobalIdGenerator,
	_captcha captcha.Captcha,
	wechat *common.WechatManager,
	authLimiter *AuthLimiter,
) *Impl {
	var authHandlers = make(map[enumsv1.AuthType]Authenticator, len(config.User.EnableAuth))
	passHasher := crypto.NewPasswordHasher(int(config.User.PasswordHasherCost))
	for _, authType := range config.User.EnableAuth {
		switch authType {
		case enumsv1.AuthType_AUTH_TYPE_NUMBER_PASSWORD:
			authHandlers[authType] = NewNumberPassword(passHasher, repo, jwtService, authLimiter)
		case enumsv1.AuthType_AUTH_TYPE_EMAIL_PASSWORD:
			authHandlers[authType] = NewEmailPassword(passHasher, repo, jwtService, authLimiter)
		case enumsv1.AuthType_AUTH_TYPE_PHONE_PASSWORD:
			authHandlers[authType] = NewPhonePassword(passHasher, repo, jwtService, authLimiter)
		case enumsv1.AuthType_AUTH_TYPE_PHONE_CAPTCHA:
			authHandlers[authType] = NewPhoneCaptcha(_captcha, repo, jwtService, shortIDGen, globalIDGen, passHasher)
		case enumsv1.AuthType_AUTH_TYPE_EMAIL_CAPTCHA:
			authHandlers[authType] = NewEmailCaptcha(_captcha, repo, jwtService, shortIDGen, globalIDGen, passHasher)
		case enumsv1.AuthType_AUTH_TYPE_WECHAT:
			authHandlers[authType] = NewWeChat(repo, jwtService, wechat, shortIDGen, globalIDGen)
		default:
			panic(fmt.Sprintf("unknown auth type: %s", authType))
		}
	}
	return &Impl{
		config:                         config,
		authHandlers:                   authHandlers,
		repo:                           repo,
		query:                          query,
		passHasher:                     passHasher,
		jwtService:                     jwtService,
		tokenVerifier:                  tokenVerifier,
		captcha:                        _captcha,
		shortIDGen:                     shortIDGen,
		globalIDGen:                    globalIDGen,
		UnimplementedUserServiceServer: userv1.UnimplementedUserServiceServer{},
	}
}

func (i *Impl) GetConfig() *configv1.Config {
	return i.config
}

func (i *Impl) CreateUser(ctx context.Context, req *userv1.CreateUserReq) (resp *userv1.CreateUserResp, err error) {
	if err = checkUserBasicUnique(ctx, i.query, i.repo, req.Biz, req.Phone, req.Number, req.Email); err != nil {
		return nil, err
	}
	userBasic, err := createUserReqToUserBasic(req, i.globalIDGen, i.shortIDGen, i.passHasher)
	if err != nil {
		return nil, err
	}
	if err = i.repo.CreateUserBasic(ctx, i.query, userBasic); err != nil {
		return nil, err
	}
	resp = &userv1.CreateUserResp{
		Uid: userBasic.ID,
	}
	return
}

func (i *Impl) UpdateUser(ctx context.Context, req *userv1.UpdateUserReq) (resp *userv1.UpdateUserResp, err error) {
	if err = checkUserBasicUnique(ctx, i.query, i.repo, req.Biz, req.Phone, req.Number, req.Email); err != nil {
		return nil, err
	}
	userBasic, err := updateUserReqToUserBasic(req, i.passHasher)
	if err != nil {
		return nil, err
	}
	if err = i.repo.UpdateUserBasicByUid(ctx, i.query, userBasic); err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *Impl) DeleteUser(ctx context.Context, req *userv1.DeleteUserReq) (resp *userv1.DeleteUserResp, err error) {
	err = i.query.Transaction(func(tx *query.Query) error {
		if err := i.repo.DeleteUserBasic(ctx, tx, req.Uid); err != nil {
			return err
		}
		if err = i.repo.DeleteUserAuth(ctx, tx, req.Uid); err != nil {
			return err
		}
		return i.revokeSessions(ctx, tx, req.Uid, enumsv1.SessionStatus_SESSION_STATUS_DELETE_OUT, "")
	})
	return nil, err
}

func (i *Impl) DeleteUsers(ctx context.Context, req *userv1.DeleteUsersReq) (resp *userv1.DeleteUsersResp, err error) {
	err = i.query.Transaction(func(tx *query.Query) error {
		if err := i.repo.DeleteUsersBasic(ctx, tx, req.Uids); err != nil {
			return err
		}
		if err = i.repo.DeleteUsersAuth(ctx, tx, req.Uids); err != nil {
			return err
		}
		g := new(errgroup.Group)
		for _, uid := range req.Uids {
			g.Go(func() error {
				return i.revokeSessions(ctx, tx, uid, enumsv1.SessionStatus_SESSION_STATUS_DELETE_OUT, "")
			})
		}
		if err = g.Wait(); err != nil {
			return err
		}
		return i.repo.DeleteUsersSignLogs(ctx, tx, req.Uids)
	})
	return nil, err
}

func (i *Impl) GetUserAuth(ctx context.Context, req *userv1.GetUserAuthReq) (resp *userv1.GetUserAuthResp, err error) {
	auth, err := i.repo.GetUserAuthByUid(ctx, i.query, req.Uid)
	if err != nil {
		return nil, err
	}
	var userAuth = make([]*userv1.UserAuth, len(auth))
	for i, a := range auth {
		userAuth[i] = userAuthModelToAuth(a)
	}
	resp = &userv1.GetUserAuthResp{
		Auth: userAuth,
	}
	return
}

func (i *Impl) UnbindUserAuth(ctx context.Context, req *userv1.UnbindUserAuthReq) (resp *userv1.UnbindUserAuthResp, err error) {
	err = i.repo.UnbindUserAuth(ctx, i.query, req.Uid, req.Identifier)
	return
}

// BindUserAuth
// 绑定第三方登录，手机，邮箱等
func (i *Impl) BindUserAuth(ctx context.Context, req *userv1.BindUserAuthReq) (resp *userv1.BindUserAuthResp, err error) {
	authenticator, err := i.getAuthenticator(req.Type)
	if err != nil {
		return nil, err
	}
	return authenticator.Bind(ctx, i.query, req)
}

func (i *Impl) SendCaptcha(ctx context.Context, req *userv1.SendCaptchaReq) (resp *userv1.SendCaptchaResp, err error) {
	res, err := i.captcha.SendCaptcha(ctx, req.CaptchaParams)
	if err != nil {
		return nil, err
	}
	resp = &userv1.SendCaptchaResp{
		Token: res.Token,
		Limit: res.Limit,
	}
	return
}

// VerifyCaptcha 验证验证码，类似通过手机找回密码，本接口会返回一个token，这个token是手机验证通过后的凭证
func (i *Impl) VerifyCaptcha(ctx context.Context, req *userv1.VerifyCaptchaReq) (resp *userv1.VerifyCaptchaResp, err error) {
	if _, err = i.captcha.VerifyCaptcha(ctx, req.Param); err != nil {
		return nil, err
	}
	var uid int64
	switch req.Param.SenderType {
	case enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS:
		uid, err = i.repo.GetUidByPhone(ctx, i.query, req.Biz, req.Param.GetPhoneNumber())
	case enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL:
		uid, err = i.repo.GetUidByEmail(ctx, i.query, req.Biz, req.Param.GetEmail())
	default:
		return nil, ecode.ErrParams
	}
	token := idx.UUIDv4()
	if err = i.tokenVerifier.Store(ctx, token, uid, req.Param.SenderType, req.Param.CaptchaType); err != nil {
		return nil, err
	}
	resp = &userv1.VerifyCaptchaResp{
		Token: token,
	}
	return
}

func (i *Impl) ChangeUserStatus(ctx context.Context, req *userv1.ChangeUserStatusReq) (resp *userv1.ChangeUserStatusResp, err error) {
	err = i.query.Transaction(func(tx *query.Query) error {
		if err = i.repo.UpdateUserBasicByUid(ctx, tx, &model.UserBasic{
			ID:     req.Uid,
			Status: int16(req.Status),
		}); err != nil {
			return err
		}
		return i.revokeAccessSessions(ctx, tx, req.Uid)
	})
	return nil, err
}

func (i *Impl) ChangePassword(ctx context.Context, req *userv1.ChangePasswordReq) (resp *userv1.ChangePasswordResp, err error) {
	err = i.query.Transaction(func(tx *query.Query) error {
		userBasic, err := i.repo.GetUserBasicByUID(ctx, tx, req.Uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ecode.ErrUserNotExist
			}
			return err
		}
		if userBasic.Password == nil {
			return ecode.ErrUserPasswordNotSet
		}
		ok, err := i.passHasher.VerifyPassword(req.OldPassword, *userBasic.Password)
		if err != nil {
			return err
		}
		if !ok {
			return ecode.ErrUserPasswordMisMatch
		}
		if err = i.updatePassword(ctx, tx, req.NewPassword, userBasic); err != nil {
			return err
		}
		return i.revokeSessions(ctx, tx, userBasic.ID, enumsv1.SessionStatus_SESSION_STATUS_CHANGE_PASSWORD_OUT, req.CurrentSessionId)
	})
	return nil, err
}

func (i *Impl) ResetPassword(ctx context.Context, req *userv1.ResetPasswordReq) (resp *userv1.ResetPasswordResp, err error) {
	uid, err := i.tokenVerifier.Verify(ctx, req.ResetToken, req.CaptchaSender, req.CaptchaType)
	if err != nil {
		return nil, err
	}
	err = i.query.Transaction(func(tx *query.Query) error {
		userBasic, err := tx.UserBasic.WithContext(ctx).Where(tx.UserBasic.ID.Eq(uid)).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ecode.ErrUserNotExist
			}
			return err
		}
		if err = i.updatePassword(ctx, tx, req.NewPassword, userBasic); err != nil {
			return err
		}
		if err = i.revokeSessions(ctx, tx, uid, enumsv1.SessionStatus_SESSION_STATUS_RESET_PASSWORD_OUT, ""); err != nil {
			return err
		}
		return i.tokenVerifier.Delete(ctx, req.ResetToken, req.CaptchaSender, req.CaptchaType)
	})
	return nil, err
}

func (i *Impl) ChangePhone(ctx context.Context, req *userv1.ChangePhoneReq) (resp *userv1.ChangePhoneResp, err error) {
	if _, err = i.captcha.VerifyCaptcha(ctx, &captchav1.VerifyCaptchaReq{
		SenderType:  enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS,
		Token:       req.CaptchaToken,
		Captcha:     req.Captcha,
		CaptchaType: req.CaptchaType,
		Number:      &captchav1.VerifyCaptchaReq_PhoneNumber{PhoneNumber: req.Phone},
	}); err != nil {
		return nil, err
	}
	uid, err := i.tokenVerifier.Verify(ctx, req.ChangeToken, enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS, req.CaptchaType)
	if err != nil {
		return nil, err
	}
	err = i.query.Transaction(func(tx *query.Query) error {
		if err = i.repo.UpdateUserBasicByUid(ctx, tx, &model.UserBasic{
			ID:               uid,
			PhoneCountryCode: req.Phone.CountryCode,
			Phone:            req.Phone.Number,
		}); err != nil {
			return err
		}
		return i.tokenVerifier.Delete(ctx, req.ChangeToken, enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS, req.CaptchaType)
	})
	return nil, err
}

func (i *Impl) ChangeEmail(ctx context.Context, req *userv1.ChangeEmailReq) (resp *userv1.ChangeEmailResp, err error) {
	if _, err = i.captcha.VerifyCaptcha(ctx, &captchav1.VerifyCaptchaReq{
		SenderType: req.Sender,
		Token:      req.CaptchaToken,
		Captcha:    req.Captcha,
		Number:     &captchav1.VerifyCaptchaReq_Email{Email: req.NewEmail},
	}); err != nil {
		return nil, err
	}
	uid, err := i.tokenVerifier.Verify(ctx, req.ChangeToken, req.Sender, req.Captcha)
	if err != nil {
		return nil, err
	}
	err = i.query.Transaction(func(tx *query.Query) error {
		if err = i.repo.UpdateUserBasicByUid(ctx, tx, &model.UserBasic{
			ID:    uid,
			Email: req.NewEmail,
		}); err != nil {
			return err
		}
		return i.tokenVerifier.Delete(ctx, req.ChangeToken, req.Sender, req.Captcha)
	})
	return nil, err
}

func (i *Impl) ChangeUserAvatar(ctx context.Context, req *userv1.ChangeUserAvatarReq) (resp *userv1.ChangeUserAvatarResp, err error) {
	err = i.repo.UpdateUserBasicByUid(ctx, i.query, &model.UserBasic{
		ID:     req.Uid,
		Avatar: trans.String(req.Avatar),
	})
	return nil, err
}

func (i *Impl) ChangeUserGender(ctx context.Context, req *userv1.ChangeUserGenderReq) (resp *userv1.ChangeUserGenderResp, err error) {
	err = i.repo.UpdateUserBasicByUid(ctx, i.query, &model.UserBasic{
		ID:     req.Uid,
		Gender: trans.Int16(int16(req.Gender)),
	})
	return nil, err
}

func (i *Impl) ChangeUserBirthday(ctx context.Context, req *userv1.ChangeUserBirthdayReq) (resp *userv1.ChangeUserBirthdayResp, err error) {
	err = i.repo.UpdateUserBasicByUid(ctx, i.query, &model.UserBasic{
		ID:       req.Uid,
		Birthday: trans.Ref(time.Date(int(req.Birthday.Year), time.Month(req.Birthday.Month), int(req.Birthday.Day), 0, 0, 0, 0, time.UTC)),
	})
	return nil, err
}

func (i *Impl) ChangeUserName(ctx context.Context, req *userv1.ChangeUserNameReq) (resp *userv1.ChangeUserNameResp, err error) {
	err = i.repo.UpdateUserBasicByUid(ctx, i.query, &model.UserBasic{
		ID:   req.Uid,
		Name: trans.String(req.Name),
	})
	return nil, err
}

func (i *Impl) ChangeUserAlias(ctx context.Context, req *userv1.ChangeUserAliasReq) (resp *userv1.ChangeUserAliasResp, err error) {
	err = i.repo.UpdateUserBasicByUid(ctx, i.query, &model.UserBasic{
		ID:     req.Uid,
		Alias_: trans.String(req.Alias),
	})
	return nil, err
}

func (i *Impl) ChangeUserNumber(ctx context.Context, req *userv1.ChangeUserNumberReq) (resp *userv1.ChangeUserNumberResp, err error) {
	err = i.query.Transaction(func(tx *query.Query) error {
		exist, err := i.repo.CheckUserNumberExists(ctx, tx, req.Number)
		if err != nil {
			return err
		}
		if exist {
			return ecode.ErrUserNumberExists
		}
		return i.repo.UpdateUserBasicByUid(ctx, tx, &model.UserBasic{
			ID:     req.Uid,
			Number: req.Number,
		})
	})
	return nil, err
}

func (i *Impl) ChangeUserAddress(ctx context.Context, req *userv1.ChangeUserAddressReq) (resp *userv1.ChangeUserAddressResp, err error) {
	err = i.repo.UpdateUserBasicByUid(ctx, i.query, &model.UserBasic{
		ID:           req.Uid,
		CountryCode:  req.CountryCode,
		ProvinceCode: req.ProvinceCode,
		CityCode:     req.CityCode,
		DistrictCode: req.DistrictCode,
		Addr:         trans.String(req.Addr),
	})
	return nil, err
}

func (i *Impl) ChangeUserType(ctx context.Context, req *userv1.ChangeUserTypeReq) (resp *userv1.ChangeUserTypeResp, err error) {
	err = i.query.Transaction(func(tx *query.Query) error {
		if err = i.repo.UpdateUserBasicByUid(ctx, tx, &model.UserBasic{
			ID:   req.Uid,
			Type: trans.Int16(int16(req.Type)),
		}); err != nil {
			return err
		}
		return i.revokeAccessSessions(ctx, tx, req.Uid)
	})
	return nil, err
}

func (i *Impl) ChangeUserLevel(ctx context.Context, req *userv1.ChangeUserLevelReq) (resp *userv1.ChangeUserLevelResp, err error) {
	err = i.query.Transaction(func(tx *query.Query) error {
		if err = i.repo.UpdateUserBasicByUid(ctx, tx, &model.UserBasic{
			ID:    req.Uid,
			Level: trans.Int32(req.Level),
		}); err != nil {
			return err
		}
		return i.revokeAccessSessions(ctx, tx, req.Uid)
	})
	return nil, err
}

func (i *Impl) ChangeUserExt(ctx context.Context, req *userv1.ChangeUserExtReq) (resp *userv1.ChangeUserExtResp, err error) {
	err = i.query.Transaction(func(tx *query.Query) error {
		if err = i.repo.UpdateUserBasicByUid(ctx, tx, &model.UserBasic{
			ID:  req.Uid,
			Ext: trans.String(req.Ext),
		}); err != nil {
			return err
		}
		return i.revokeAccessSessions(ctx, tx, req.Uid)
	})
	return nil, err
}

func (i *Impl) VerifyPhone(ctx context.Context, req *userv1.VerifyPhoneReq) (resp *userv1.VerifyPhoneResp, err error) {
	if _, err = i.captcha.VerifyCaptcha(ctx, &captchav1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS,
		Token:      req.CaptchaToken,
		Captcha:    req.Captcha,
		Number:     &captchav1.VerifyCaptchaReq_PhoneNumber{PhoneNumber: req.Phone},
	}); err != nil {
		return nil, err
	}
	userBasic, err := i.repo.GetUserBasicByUID(ctx, i.query, req.Uid)
	if err != nil {
		return nil, err
	}
	if userBasic.PhoneCountryCode != req.Phone.CountryCode || userBasic.Phone != req.Phone.Number {
		return nil, ecode.ErrPhoneNotMatch
	}
	err = i.repo.UpdateUserBasicByUid(ctx, i.query, &model.UserBasic{
		ID:            req.Uid,
		PhoneVerified: int16(enumsv1.Verified_VERIFIED_VERIFIED),
	})
	return nil, err
}

func (i *Impl) VerifyEmail(ctx context.Context, req *userv1.VerifyEmailReq) (resp *userv1.VerifyEmailResp, err error) {
	if _, err = i.captcha.VerifyCaptcha(ctx, &captchav1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL,
		Token:      req.CaptchaToken,
		Captcha:    req.Captcha,
		Number:     &captchav1.VerifyCaptchaReq_Email{Email: req.Email},
	}); err != nil {
		return nil, err
	}
	userBasic, err := i.repo.GetUserBasicByUID(ctx, i.query, req.Uid)
	if err != nil {
		return nil, err
	}
	if userBasic.Email != req.Email {
		return nil, ecode.ErrEmailNotMatch
	}
	err = i.repo.UpdateUserBasicByUid(ctx, i.query, &model.UserBasic{
		ID:            req.Uid,
		EmailVerified: int16(enumsv1.Verified_VERIFIED_VERIFIED),
	})
	return nil, err
}

func (i *Impl) SignUp(ctx context.Context, req *userv1.SignUpReq) (resp *userv1.SignUpResp, err error) {
	authenticator, err := i.getAuthenticator(req.AuthType)
	if err != nil {
		return nil, err
	}
	return authenticator.SignUp(ctx, i.query, req)
}

func (i *Impl) SignIn(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	authenticator, err := i.getAuthenticator(req.AuthType)
	if err != nil {
		return nil, err
	}
	return authenticator.SignIn(ctx, i.query, req)
}

func (i *Impl) SignOut(ctx context.Context, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error) {
	authenticator, err := i.getAuthenticator(req.AuthType)
	if err != nil {
		return nil, err
	}
	return authenticator.SignOut(ctx, i.query, req)
}

func (i *Impl) SignOutByUid(ctx context.Context, req *userv1.SignOutByUidReq) (resp *userv1.SignOutByUidResp, err error) {
	logs, err := i.repo.GetActiveSignInLogs(ctx, i.query, req.Uid)
	if err != nil {
		return nil, err
	}
	g := new(errgroup.Group)
	for _, log := range logs {
		l := log
		t := enumsv1.AuthType(log.Type)
		g.Go(func() error {
			authenticator, err := i.getAuthenticator(t)
			if err != nil {
				return err
			}
			_, err = authenticator.SignOut(ctx, i.query, &userv1.SignOutReq{
				SessionId: l.AccessSessionID,
				Reason:    req.Reason,
				AuthType:  t,
			})
			return err
		})
	}
	err = g.Wait()
	return nil, err
}

func (i *Impl) Refresh(ctx context.Context, req *userv1.RefreshReq) (resp *userv1.RefreshResp, err error) {
	access, refresh, _, _, err := i.jwtService.RefreshToken(ctx, i.query, req.RefreshToken, req.ExtraJwtClaims)
	if err != nil {
		return nil, err
	}
	resp = &userv1.RefreshResp{
		NewAccessToken:  access,
		NewRefreshToken: refresh,
	}
	return resp, nil
}

func (i *Impl) VerifyToken(ctx context.Context, req *userv1.VerifyTokenReq) (resp *userv1.VerifyTokenResp, err error) {
	var claims *JwtClaims
	if req.Type == enumsv1.TokenType_TOKEN_TYPE_ACCESS {
		claims, err = i.jwtService.VerifyAccessToken(ctx, req.Token)
	} else if req.Type == enumsv1.TokenType_TOKEN_TYPE_REFRESH {
		claims, err = i.jwtService.VerifyRefreshToken(ctx, req.Token)
	} else {
		return nil, ecode.ErrParams
	}
	if err != nil {
		return nil, err
	}
	resp = &userv1.VerifyTokenResp{
		UserInfo: &userv1.UserProfile{
			Uid:       claims.Uid,
			Number:    claims.Number,
			Biz:       claims.Biz,
			UserType:  claims.UserType,
			UserLevel: claims.UserLevel,
			AuthType:  enumsv1.AuthType(claims.AuthType),
			Appid:     claims.AppId,
			Openid:    claims.OpenId,
			Unionid:   claims.UnionId,
			SessionId: claims.ID,
			Extra:     claims.Extra,
		},
	}
	return resp, nil
}

func (i *Impl) GetActiveSignInLogs(ctx context.Context, req *userv1.GetActiveSignInLogsReq) (resp *userv1.GetActiveSignInLogsResp, err error) {
	activeLogs, err := i.repo.GetActiveSignInLogs(ctx, i.query, req.Uid)
	if err != nil {
		return nil, err
	}
	var logs = make([]*userv1.SignInLog, len(activeLogs))
	for i, log := range activeLogs {
		logs[i] = userSignInModelToSignInLog(log)
	}
	resp = &userv1.GetActiveSignInLogsResp{
		Logs: logs,
	}
	return resp, nil
}

func (i *Impl) PagingGetSignInLogs(ctx context.Context, req *userv1.PagingGetSignInLogsReq) (resp *userv1.PagingGetSignInLogsResp, err error) {
	res, err := i.repo.PagingGetSignInLogs(ctx, i.query, req)
	if err != nil {
		return nil, err
	}
	var logs = make([]*userv1.SignInLog, len(res.List))
	for i, log := range res.List {
		logs[i] = userSignInModelToSignInLog(log)
	}
	resp = &userv1.PagingGetSignInLogsResp{
		Logs:       logs,
		Page:       res.Page,
		Size:       res.PageSize,
		Total:      res.Total,
		TotalPages: res.TotalPages,
	}
	return resp, nil
}

func (i *Impl) PagingGetUsers(ctx context.Context, req *userv1.PagingGetUsersReq) (resp *userv1.PagingGetUsersResp, err error) {
	res, err := i.repo.PagingGetUsers(ctx, i.query, req)
	if err != nil {
		return nil, err
	}
	var users = make([]*userv1.UserInfo, len(res.List))
	for i, user := range res.List {
		users[i] = userBasicToUserInfo(user)
	}
	resp = &userv1.PagingGetUsersResp{
		Users:      users,
		Page:       res.Page,
		Size:       res.PageSize,
		Total:      res.Total,
		TotalPages: res.TotalPages,
	}
	return resp, nil
}

func (i *Impl) getAuthenticator(authType enumsv1.AuthType) (Authenticator, error) {
	authenticator, ok := i.authHandlers[authType]
	if !ok {
		return nil, ecode.ErrUnImplemented
	}
	return authenticator, nil
}

func (i *Impl) updatePassword(ctx context.Context, tx *query.Query, password string, userBasic *model.UserBasic) error {
	newPassword, err := i.passHasher.HashPassword(password)
	if err != nil {
		return err
	}
	userBasic.Password = &newPassword
	userBasic.PasswordUpdatedAt = time.Now().UnixMilli()
	return i.repo.UpdateUserBasicByUid(ctx, tx, userBasic)
}

func (i *Impl) revokeAccessSessions(ctx context.Context, tx *query.Query, uid int64) error {
	activeLogs, err := i.repo.GetActiveSignInLogs(ctx, tx, uid)
	if err != nil {
		return err
	}
	if len(activeLogs) == 0 {
		return nil
	}
	var items []*BlockItem
	for _, item := range activeLogs {
		items = append(items, &BlockItem{
			Target: item.AccessSessionID,
			TTL:    i.jwtService.TTLFromExpiredAt(item.AccessExpiredAt),
		})
	}
	return i.jwtService.RevokeTokens(ctx, items)
}

func (i *Impl) revokeSessions(ctx context.Context, tx *query.Query, uid int64, reason enumsv1.SessionStatus, exceptSessionId string) error {
	activeLogs, err := i.repo.GetActiveSignInLogs(ctx, tx, uid)
	if len(activeLogs) == 0 {
		return nil
	}
	var ids []int64
	var items []*BlockItem
	for _, activeLog := range activeLogs {
		if activeLog.AccessSessionID == exceptSessionId {
			continue
		}
		ids = append(ids, activeLog.ID)
		items = append(items,
			&BlockItem{
				Target: activeLog.AccessSessionID,
				TTL:    i.jwtService.TTLFromExpiredAt(activeLog.AccessExpiredAt),
			},
			&BlockItem{
				Target: activeLog.RefreshSessionID,
				TTL:    i.jwtService.TTLFromExpiredAt(activeLog.RefreshExpiredAt),
			},
		)
	}
	if _, err = tx.UserSignLog.WithContext(ctx).Where(tx.UserSignLog.ID.In(ids...)).Update(tx.UserSignLog.Status, int16(reason)); err != nil {
		return err
	}
	return i.jwtService.RevokeTokens(ctx, items)
}
