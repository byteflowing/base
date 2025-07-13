package user

import (
	"context"
	"github.com/byteflowing/base/kitex_gen/base"
)

type User interface {
	SendPhoneCaptcha(ctx context.Context, req *base.SendPhoneCaptchaReq) (resp *base.SendPhoneCaptchaResp, err error)
	SendEmailCaptcha(ctx context.Context, req *base.SendEmailCaptchaReq) (resp *base.SendEmailCaptchaResp, err error)
	Login(ctx context.Context, req *base.LoginReq) (resp *base.LoginResp, err error)
	Logout(ctx context.Context, req *base.LogoutReq) (resp *base.LogoutResp, err error)
	ForceLogoutBySessionId(ctx context.Context, req *base.ForceLogoutBySessionIdReq) (resp *base.LogoutResp, err error)
	RefreshToken(ctx context.Context, req *base.RefreshTokenReq) (resp *base.RefreshTokenResp, err error)
	VerifyToken(ctx context.Context, req *base.VerifyTokenReq) (resp *base.VerifyTokenResp, err error)
	AddUser(ctx context.Context, req *base.AddUserReq) (resp *base.AddUserResp, err error)
	BatchAddUsers(ctx context.Context, req *base.BatchAddUsersReq) (resp *base.BatchAddUsersResp, err error)
	UpdateUser(ctx context.Context, req *base.UpdateUserReq) (resp *base.UpdateUserResp, err error)
	DeleteUser(ctx context.Context, req *base.DeleteUserReq) (resp *base.DeleteUserResp, err error)
	BatchDeleteUsers(ctx context.Context, req *base.BatchDeleteUsersReq) (resp *base.BatchDeleteUsersResp, err error)
	GetUserByNumber(ctx context.Context, req *base.GetUserByNumberReq) (resp *base.GetUserByNumberResp, err error)
	PagingGetUsers(ctx context.Context, req *base.PagingGetUsersReq) (resp *base.PagingGetUsersResp, err error)
	GetUserLoginLogs(ctx context.Context, req *base.GetUserLoginLogsReq) (resp *base.GetUserLoginLogsResp, err error)
	PagingGetLoginLogs(ctx context.Context, req *base.PagingGetLoginLogsReq) (resp *base.PagingGetLoginLogsResp, err error)
}
