package user

import (
	"context"

	"github.com/byteflowing/base/biz/constant"
	"github.com/byteflowing/base/kitex_gen/base"
)

type EmptyImpl struct{}

func (e *EmptyImpl) SendLoginCaptcha(ctx context.Context, req *base.SendLoginCaptchaReq) (res *base.SendLoginCaptchaResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) Login(ctx context.Context, req *base.LoginReq) (resp *base.LoginResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) Logout(ctx context.Context, req *base.LogoutReq) (resp *base.LogoutResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) ForceLogoutBySessionId(ctx context.Context, req *base.ForceLogoutBySessionIdReq) (resp *base.LogoutResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) RefreshToken(ctx context.Context, req *base.RefreshTokenReq) (resp *base.RefreshTokenResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) VerifyToken(ctx context.Context, req *base.VerifyTokenReq) (resp *base.VerifyTokenResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) AddUser(ctx context.Context, req *base.AddUserReq) (resp *base.AddUserResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) BatchAddUsers(ctx context.Context, req *base.BatchAddUsersReq) (resp *base.BatchAddUsersResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) UpdateUser(ctx context.Context, req *base.UpdateUserReq) (resp *base.UpdateUserResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) DeleteUser(ctx context.Context, req *base.DeleteUserReq) (resp *base.DeleteUserResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) BatchDeleteUsers(ctx context.Context, req *base.BatchDeleteUsersReq) (resp *base.BatchDeleteUsersResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) GetUserByNumber(ctx context.Context, req *base.GetUserByNumberReq) (resp *base.GetUserByNumberResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) PagingGetUsers(ctx context.Context, req *base.PagingGetUsersReq) (resp *base.PagingGetUsersResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) GetUserLoginLogs(ctx context.Context, req *base.GetUserLoginLogsReq) (resp *base.GetUserLoginLogsResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) PagingGetLoginLogs(ctx context.Context, req *base.PagingGetLoginLogsReq) (resp *base.PagingGetLoginLogsResp, err error) {
	return nil, constant.ErrNotActive
}
