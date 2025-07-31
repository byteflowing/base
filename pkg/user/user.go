package user

import (
	"context"
	"github.com/byteflowing/base/dal/model"

	"github.com/byteflowing/base/dal/query"
)

type Authenticator interface {
	AuthType() Type
	Authenticate(ctx context.Context, req *LoginReq) (resp *LoginResp, err error)
}

type User interface {
	Login(ctx context.Context, req *LoginReq) (resp *LoginResp, err error)
	Logout(ctx context.Context) (err error)
	LogoutBySessionId(ctx context.Context, sessionId string) (err error)
	Refresh(ctx context.Context, refreshToken string) (newToken string, err error)
	VerifyToken(ctx context.Context, token string) (err error)
	GetOnlineLoginLog(ctx context.Context, uid uint64) (logs []*model.UserLoginLog, err error)
	PagingGetLoginLogs(ctx context.Context, req *PagingGetLoginLogsReq) (resp *PagingGetLoginLogsResp, err error)
}

type Impl struct {
	handlers map[Type]Authenticator
}

func NewUserService(db *query.Query) User {
	return &Impl{}
}

func (i *Impl) Login(ctx context.Context, req *LoginReq) (resp *LoginResp, err error) {
	return nil, nil
}

func (i *Impl) Logout(ctx context.Context) (err error) {
	return nil
}

func (i *Impl) LogoutBySessionId(ctx context.Context, sessionId string) (err error) {
	return nil
}

func (i *Impl) Refresh(ctx context.Context, refreshToken string) (newToken string, err error) {
	return "", nil
}

func (i *Impl) VerifyToken(ctx context.Context, token string) (err error) {
	return nil
}

func (i *Impl) GetOnlineLoginLog(ctx context.Context, uid uint64) (logs []*model.UserLoginLog, err error) {
	return nil, nil
}

func (i *Impl) PagingGetLoginLogs(ctx context.Context, req *PagingGetLoginLogsReq) (resp *PagingGetLoginLogsResp, err error) {
	return nil, err
}
