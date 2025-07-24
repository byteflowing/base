package user

import (
	"context"

	"github.com/byteflowing/base/biz/config"
	"github.com/byteflowing/base/biz/dal/query"
	"github.com/byteflowing/base/biz/pkg/message"
	"github.com/byteflowing/base/kitex_gen/base"
	"github.com/byteflowing/go-common/redis"
)

type User interface {
	SendLoginCaptcha(ctx context.Context, req *base.SendLoginCaptchaReq) (res *base.SendLoginCaptchaResp, err error)
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

type Opts struct {
	Config  *config.UserConfig
	DB      *query.Query
	RDB     *redis.Redis
	Message message.Message
}

type Impl struct {
	conf    *config.UserConfig
	db      *query.Query
	redis   *redis.Redis
	message message.Message
}

func New(opts *Opts) User {
	if opts.Config == nil || !opts.Config.Active {
		return new(EmptyImpl)
	}
	return &Impl{
		conf:    opts.Config,
		db:      opts.DB,
		redis:   opts.RDB,
		message: opts.Message,
	}
}

func (i Impl) SendLoginCaptcha(ctx context.Context, req *base.SendLoginCaptchaReq) (res *base.SendLoginCaptchaResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) Login(ctx context.Context, req *base.LoginReq) (resp *base.LoginResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) Logout(ctx context.Context, req *base.LogoutReq) (resp *base.LogoutResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) ForceLogoutBySessionId(ctx context.Context, req *base.ForceLogoutBySessionIdReq) (resp *base.LogoutResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) RefreshToken(ctx context.Context, req *base.RefreshTokenReq) (resp *base.RefreshTokenResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) VerifyToken(ctx context.Context, req *base.VerifyTokenReq) (resp *base.VerifyTokenResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) AddUser(ctx context.Context, req *base.AddUserReq) (resp *base.AddUserResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) BatchAddUsers(ctx context.Context, req *base.BatchAddUsersReq) (resp *base.BatchAddUsersResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) UpdateUser(ctx context.Context, req *base.UpdateUserReq) (resp *base.UpdateUserResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) DeleteUser(ctx context.Context, req *base.DeleteUserReq) (resp *base.DeleteUserResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) BatchDeleteUsers(ctx context.Context, req *base.BatchDeleteUsersReq) (resp *base.BatchDeleteUsersResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) GetUserByNumber(ctx context.Context, req *base.GetUserByNumberReq) (resp *base.GetUserByNumberResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) PagingGetUsers(ctx context.Context, req *base.PagingGetUsersReq) (resp *base.PagingGetUsersResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) GetUserLoginLogs(ctx context.Context, req *base.GetUserLoginLogsReq) (resp *base.GetUserLoginLogsResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (i Impl) PagingGetLoginLogs(ctx context.Context, req *base.PagingGetLoginLogsReq) (resp *base.PagingGetLoginLogsResp, err error) {
	//TODO implement me
	panic("implement me")
}
