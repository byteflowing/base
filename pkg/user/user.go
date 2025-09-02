package user

import (
	"context"
	"errors"
	"time"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/dal/query"
	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	msgv1 "github.com/byteflowing/base/gen/msg/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/base/pkg/captcha"
	"github.com/byteflowing/go-common/crypto"
	"github.com/byteflowing/go-common/idx"
	"github.com/byteflowing/go-common/trans"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Authenticator interface {
	AuthType() enumsv1.AuthType
	SignUp(ctx context.Context, tx *query.Query, req *userv1.SignUpReq) (*userv1.SignUpResp, error)
	SignIn(ctx context.Context, tx *query.Query, req *userv1.SignInReq) (resp *userv1.SignInResp, err error)
	SignOut(ctx context.Context, tx *query.Query, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error)
}

type User interface {
	SendCaptcha(ctx context.Context, req *userv1.SendCaptchaReq) (resp *userv1.SendCaptchaResp, err error)
	VerifyCaptcha(ctx context.Context, req *userv1.VerifyCaptchaReq) (resp *userv1.VerifyCaptchaResp, err error)
	SignUp(ctx context.Context, req *userv1.SignUpReq) (resp *userv1.SignUpResp, err error)
	SignIn(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error)
	SignOut(ctx context.Context, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error)
	SignOutByUid(ctx context.Context, req *userv1.SignOutByUidReq) (resp *userv1.SignOutByUidResp, err error)
	Refresh(ctx context.Context, req *userv1.RefreshReq) (resp *userv1.RefreshResp, err error)
	ChangePassword(ctx context.Context, req *userv1.ChangePasswordReq) (resp *userv1.ChangePasswordResp, err error)
	ResetPassword(ctx context.Context, req *userv1.ResetPasswordReq) (resp *userv1.ResetPasswordResp, err error)
	ChangeUserStatus(ctx context.Context, req *userv1.ChangeUserStatusReq) (resp *userv1.ChangeUserStatusResp, err error)
	ChangePhone(ctx context.Context, req *userv1.ChangePhoneReq) (resp *userv1.ChangePhoneResp, err error)
	ChangeEmail(ctx context.Context, req *userv1.ChangeEmailReq) (resp *userv1.ChangeEmailResp, err error)
	ChangeUserAvatar(ctx context.Context, req *userv1.ChangeUserAvatarReq) (resp *userv1.ChangeUserAvatarResp, err error)
	ChangeUserGender(ctx context.Context, req *userv1.ChangeUserGenderReq) (resp *userv1.ChangeUserGenderResp, err error)
	ChangeUserBirthday(ctx context.Context, req *userv1.ChangeUserBirthdayReq) (resp *userv1.ChangeUserBirthdayResp, err error)
	ChangeUserName(ctx context.Context, req *userv1.ChangeUserNameReq) (resp *userv1.ChangeUserNameResp, err error)
	ChangeUserAlias(ctx context.Context, req *userv1.ChangeUserAliasReq) (resp *userv1.ChangeUserAliasResp, err error)
	ChangeUserNumber(ctx context.Context, req *userv1.ChangeUserNumberReq) (resp *userv1.ChangeUserNumberResp, err error)
	ChangeUserAddress(ctx context.Context, req *userv1.ChangeUserAddressReq) (resp *userv1.ChangeUserAddressResp, err error)
	ChangeUserType(ctx context.Context, req *userv1.ChangeUserTypeReq) (resp *userv1.ChangeUserTypeResp, err error)
	ChangeUserLevel(ctx context.Context, req *userv1.ChangeUserLevelReq) (resp *userv1.ChangeUserLevelResp, err error)
	ChangeUserExt(ctx context.Context, req *userv1.ChangeUserExtReq) (resp *userv1.ChangeUserExtResp, err error)
	VerifyPhone(ctx context.Context, req *userv1.VerifyPhoneReq) (resp *userv1.VerifyPhoneResp, err error)
	VerifyEmail(ctx context.Context, req *userv1.VerifyEmailReq) (resp *userv1.VerifyEmailResp, err error)
	VerifyToken(ctx context.Context, req *userv1.VerifyTokenReq) (resp *userv1.VerifyTokenResp, err error)
	GetActiveSignInLogs(ctx context.Context, req *userv1.GetActiveSignInLogsReq) (resp *userv1.GetActiveSignInLogsResp, err error)
	PagingGetSignInLogs(ctx context.Context, req *userv1.PagingGetSignInLogsReq) (resp *userv1.PagingGetSignInLogsResp, err error)
}

type Impl struct {
	authHandlers  map[enumsv1.AuthType]Authenticator
	repo          Repo
	query         *query.Query
	passHasher    *crypto.PasswordHasher
	jwtService    *JwtService
	tokenVerifier TokenVerifier
	captcha       captcha.Captcha
}

func (i *Impl) SendCaptcha(ctx context.Context, req *userv1.SendCaptchaReq) (resp *userv1.SendCaptchaResp, err error) {
	res, err := i.captcha.SendCaptcha(ctx, req.CaptchaParams)
	if err != nil {
		return nil, err
	}
	resp = &userv1.SendCaptchaResp{
		Data: &userv1.SendCaptchaResp_Data{
			Token: res.Data.Token,
			Limit: res.Data.Limit,
		},
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
	if err = i.tokenVerifier.Store(ctx, token, uid); err != nil {
		return nil, err
	}
	resp = &userv1.VerifyCaptchaResp{
		Data: &userv1.VerifyCaptchaResp_Data{Token: token},
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
	uid, err := i.tokenVerifier.GetUid(ctx, req.ResetToken)
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
		return i.revokeSessions(ctx, tx, uid, enumsv1.SessionStatus_SESSION_STATUS_RESET_PASSWORD_OUT, "")
	})
	return nil, err
}

func (i *Impl) ChangePhone(ctx context.Context, req *userv1.ChangePhoneReq) (resp *userv1.ChangePhoneResp, err error) {
	if _, err = i.captcha.VerifyCaptcha(ctx, &msgv1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS,
		Token:      req.CaptchaToken,
		Captcha:    req.Captcha,
		Number:     &msgv1.VerifyCaptchaReq_PhoneNumber{PhoneNumber: req.Phone},
	}); err != nil {
		return nil, err
	}
	uid, err := i.tokenVerifier.GetUid(ctx, req.ChangeToken)
	if err != nil {
		return nil, err
	}
	err = i.repo.UpdateUserBasicByUid(ctx, i.query, &model.UserBasic{
		ID:               uid,
		PhoneCountryCode: req.Phone.CountryCode,
		Phone:            req.Phone.Number,
	})
	return nil, err
}

func (i *Impl) ChangeEmail(ctx context.Context, req *userv1.ChangeEmailReq) (resp *userv1.ChangeEmailResp, err error) {
	if _, err = i.captcha.VerifyCaptcha(ctx, &msgv1.VerifyCaptchaReq{
		SenderType: req.Sender,
		Token:      req.CaptchaToken,
		Captcha:    req.Captcha,
		Number:     &msgv1.VerifyCaptchaReq_Email{Email: req.NewEmail},
	}); err != nil {
		return nil, err
	}
	uid, err := i.tokenVerifier.GetUid(ctx, req.ChangeToken)
	if err != nil {
		return nil, err
	}
	err = i.repo.UpdateUserBasicByUid(ctx, i.query, &model.UserBasic{
		ID:    uid,
		Email: req.NewEmail,
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
	if _, err = i.captcha.VerifyCaptcha(ctx, &msgv1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS,
		Token:      req.CaptchaToken,
		Captcha:    req.Captcha,
		Number:     &msgv1.VerifyCaptchaReq_PhoneNumber{PhoneNumber: req.Phone},
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
	if _, err = i.captcha.VerifyCaptcha(ctx, &msgv1.VerifyCaptchaReq{
		SenderType: enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL,
		Token:      req.CaptchaToken,
		Captcha:    req.Captcha,
		Number:     &msgv1.VerifyCaptchaReq_Email{Email: req.Email},
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
		Data: &userv1.RefreshResp_Data{
			NewAccessToken:  access,
			NewRefreshToken: refresh,
		},
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
		Data: &userv1.VerifyTokenResp_Data{
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
		Data: &userv1.GetActiveSignInLogsResp_Data{
			Logs: logs,
		},
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
		Data: &userv1.PagingGetSignInLogsResp_Data{
			Logs:       logs,
			Page:       res.Page,
			Size:       res.PageSize,
			Total:      res.Total,
			TotalPages: res.TotalPages,
		},
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
	var items []*SessionItem
	for _, item := range activeLogs {
		items = append(items, &SessionItem{
			SessionID: item.AccessSessionID,
			TTL:       i.jwtService.TTLFromExpiredAt(item.AccessExpiredAt),
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
	var items []*SessionItem
	for _, log := range activeLogs {
		if log.AccessSessionID == exceptSessionId {
			continue
		}
		ids = append(ids, log.ID)
		items = append(items,
			&SessionItem{
				SessionID: log.AccessSessionID,
				TTL:       i.jwtService.TTLFromExpiredAt(log.AccessExpiredAt),
			},
			&SessionItem{
				SessionID: log.RefreshSessionID,
				TTL:       i.jwtService.TTLFromExpiredAt(log.RefreshExpiredAt),
			},
		)
	}
	if _, err = tx.UserSignLog.WithContext(ctx).Where(tx.UserSignLog.ID.In(ids...)).Update(tx.UserSignLog.Status, int16(reason)); err != nil {
		return err
	}
	return i.jwtService.RevokeTokens(ctx, items)
}
