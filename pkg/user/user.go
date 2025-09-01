package user

import (
	"context"
	"errors"
	"time"

	"github.com/byteflowing/base/dal/model"
	"github.com/byteflowing/base/dal/query"
	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
	"github.com/byteflowing/base/pkg/captcha"
	"github.com/byteflowing/go-common/crypto"
	"github.com/byteflowing/go-common/idx"
	"gorm.io/gorm"
)

type Authenticator interface {
	AuthType() enumsv1.AuthType
	SignUp(ctx context.Context, req *userv1.SignUpReq) (*userv1.SignUpResp, error)
	SignIn(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error)
	SignOut(ctx context.Context, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error)
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
	ChangePhone(ctx context.Context, req *userv1.ChangePhoneReq) (resp *userv1.ChangePhoneResp, err error)
	ChangeEmail(ctx context.Context, req *userv1.ChangeEmailReq) (resp *userv1.ChangeEmailResp, err error)
	VerifyPhone(ctx context.Context, req *userv1.VerifyPhoneReq) (resp *userv1.VerifyPhoneResp, err error)
	VerifyEmail(ctx context.Context, req *userv1.VerifyEmailReq) (resp *userv1.VerifyEmailResp, err error)
	VerifyToken(ctx context.Context, req *userv1.VerifyTokenReq) (resp *userv1.VerifyTokenResp, err error)
	ChangeUserStatus(ctx context.Context, req *userv1.ChangeUserStatusReq) (resp *userv1.ChangeUserStatusResp, err error)
	GetActiveSignInLogs(ctx context.Context, req *userv1.GetActiveSignInLogsReq) (resp *userv1.GetActiveSignInLogsResp, err error)
	PagingGetSignInLogs(ctx context.Context, req *userv1.PagingGetSignInLogsReq) (resp *userv1.PagingGetSignInLogsResp, err error)
	AddSessionToBlockList(ctx context.Context, req *userv1.AddSessionToBlockListReq) (resp *userv1.AddSessionToBlockListResp, err error)
}

type Impl struct {
	authHandlers  map[enumsv1.AuthType]Authenticator
	repo          Repo
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
		uid, err = i.repo.GetUidByPhone(ctx, req.Biz, req.Param.GetPhoneNumber())
	case enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL:
		uid, err = i.repo.GetUidByEmail(ctx, req.Biz, req.Param.GetEmail())
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

func (i *Impl) ChangePhone(ctx context.Context, req *userv1.ChangePhoneReq) (resp *userv1.ChangePhoneResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) ChangeEmail(ctx context.Context, req *userv1.ChangeEmailReq) (resp *userv1.ChangeEmailResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) VerifyPhone(ctx context.Context, req *userv1.VerifyPhoneReq) (resp *userv1.VerifyPhoneResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) VerifyEmail(ctx context.Context, req *userv1.VerifyEmailReq) (resp *userv1.VerifyEmailResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) ChangeUserStatus(ctx context.Context, req *userv1.ChangeUserStatusReq) (resp *userv1.ChangeUserStatusResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) ChangePassword(ctx context.Context, req *userv1.ChangePasswordReq) (resp *userv1.ChangePasswordResp, err error) {
	err = i.repo.Transaction(func(tx *query.Query) error {
		userBasic, err := tx.UserBasic.WithContext(ctx).Where(tx.UserBasic.ID.Eq(req.Uid)).Take()
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
	err = i.repo.Transaction(func(tx *query.Query) error {
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

func (i *Impl) SignUp(ctx context.Context, req *userv1.SignUpReq) (resp *userv1.SignUpResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) SignIn(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) SignOut(ctx context.Context, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) SignOutByUid(ctx context.Context, req *userv1.SignOutByUidReq) (resp *userv1.SignOutByUidResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) Refresh(ctx context.Context, req *userv1.RefreshReq) (resp *userv1.RefreshResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) VerifyToken(ctx context.Context, req *userv1.VerifyTokenReq) (resp *userv1.VerifyTokenResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) GetActiveSignInLogs(ctx context.Context, req *userv1.GetActiveSignInLogsReq) (resp *userv1.GetActiveSignInLogsResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) PagingGetSignInLogs(ctx context.Context, req *userv1.PagingGetSignInLogsReq) (resp *userv1.PagingGetSignInLogsResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) AddSessionToBlockList(ctx context.Context, req *userv1.AddSessionToBlockListReq) (resp *userv1.AddSessionToBlockListResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) updatePassword(ctx context.Context, tx *query.Query, password string, userBasic *model.UserBasic) error {
	newPassword, err := i.passHasher.HashPassword(password)
	if err != nil {
		return err
	}
	userBasic.Password = &newPassword
	userBasic.PasswordUpdatedAt = time.Now().UnixMilli()
	_, err = tx.UserBasic.WithContext(ctx).Where(tx.UserBasic.ID.Eq(userBasic.ID)).Updates(userBasic)
	return err
}

func (i *Impl) revokeSessions(ctx context.Context, tx *query.Query, uid int64, reason enumsv1.SessionStatus, exceptSessionId string) error {
	now := time.Now().UnixMilli()
	activeLogs, err := tx.UserSignLog.WithContext(ctx).Where(
		tx.UserSignLog.UID.Eq(int64(uid)),
		tx.UserBasic.Status.Eq(int16(enumsv1.SessionStatus_SESSION_STATUS_OK)),
		tx.UserSignLog.RefreshExpiredAt.Gt(now),
	).Order(tx.UserSignLog.RefreshExpiredAt.Desc()).Find()
	if err != nil {
		return err
	}
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
