package main

import (
	"context"
	base "github.com/byteflowing/base/kitex_gen/base"
)

// BaseServiceImpl implements the last service interface defined in the IDL.
type BaseServiceImpl struct{}

// SendCaptcha implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) SendCaptcha(ctx context.Context, req *base.SendCaptchaReq) (resp *base.SendCaptchaResp, err error) {
	// TODO: Your code here...
	return
}

// VerifyCaptcha implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) VerifyCaptcha(ctx context.Context, req *base.VerifyCaptchaReq) (resp *base.VerifyCaptchaResp, err error) {
	// TODO: Your code here...
	return
}

// PagingGetSmsMessages implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) PagingGetSmsMessages(ctx context.Context, req *base.PagingGetSmsMessagesReq) (resp *base.PagingGetSmsMessagesResp, err error) {
	// TODO: Your code here...
	return
}

// SendLoginCaptcha implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) SendLoginCaptcha(ctx context.Context, req *base.SendLoginCaptchaReq) (resp *base.SendLoginCaptchaResp, err error) {
	// TODO: Your code here...
	return
}

// Login implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) Login(ctx context.Context, req *base.LoginReq) (resp *base.LoginResp, err error) {
	// TODO: Your code here...
	return
}

// Logout implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) Logout(ctx context.Context, req *base.LogoutReq) (resp *base.LogoutResp, err error) {
	// TODO: Your code here...
	return
}

// ForceLogoutBySessionId implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) ForceLogoutBySessionId(ctx context.Context, req *base.ForceLogoutBySessionIdReq) (resp *base.ForceLogoutBySessionIdResp, err error) {
	// TODO: Your code here...
	return
}

// RefreshToken implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) RefreshToken(ctx context.Context, req *base.RefreshTokenReq) (resp *base.RefreshTokenResp, err error) {
	// TODO: Your code here...
	return
}

// VerifyToken implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) VerifyToken(ctx context.Context, req *base.VerifyTokenReq) (resp *base.VerifyTokenResp, err error) {
	// TODO: Your code here...
	return
}

// AddUser implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) AddUser(ctx context.Context, req *base.AddUserReq) (resp *base.AddUserResp, err error) {
	// TODO: Your code here...
	return
}

// BatchAddUsers implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) BatchAddUsers(ctx context.Context, req *base.BatchAddUsersReq) (resp *base.BatchAddUsersResp, err error) {
	// TODO: Your code here...
	return
}

// UpdateUser implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) UpdateUser(ctx context.Context, req *base.UpdateUserReq) (resp *base.UpdateUserResp, err error) {
	// TODO: Your code here...
	return
}

// DeleteUser implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) DeleteUser(ctx context.Context, req *base.DeleteUserReq) (resp *base.DeleteUserResp, err error) {
	// TODO: Your code here...
	return
}

// BatchDeleteUsers implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) BatchDeleteUsers(ctx context.Context, req *base.BatchDeleteUsersReq) (resp *base.BatchDeleteUsersResp, err error) {
	// TODO: Your code here...
	return
}

// GetUserByNumber implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) GetUserByNumber(ctx context.Context, req *base.GetUserByNumberReq) (resp *base.GetUserByNumberResp, err error) {
	// TODO: Your code here...
	return
}

// PagingGetUsers implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) PagingGetUsers(ctx context.Context, req *base.PagingGetUsersReq) (resp *base.PagingGetUsersResp, err error) {
	// TODO: Your code here...
	return
}

// GetUserLoginLogs implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) GetUserLoginLogs(ctx context.Context, req *base.GetUserLoginLogsReq) (resp *base.GetUserLoginLogsResp, err error) {
	// TODO: Your code here...
	return
}

// PagingGetLoginLogs implements the BaseServiceImpl interface.
func (s *BaseServiceImpl) PagingGetLoginLogs(ctx context.Context, req *base.PagingGetLoginLogsReq) (resp *base.PagingGetLoginLogsResp, err error) {
	// TODO: Your code here...
	return
}
