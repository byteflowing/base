package user

import (
	"context"

	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	userv1 "github.com/byteflowing/base/gen/user/v1"
)

type Authenticator interface {
	AuthType() enumsv1.AuthType
	SignUp(ctx context.Context, req *userv1.SignUpReq) (*userv1.SignUpResp, error)
	SignIn(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error)
	SignOut(ctx context.Context, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error)
}

type User interface {
	SignUp(ctx context.Context, req *userv1.SignUpReq) (resp *userv1.SignUpResp, err error)
	SignIn(ctx context.Context, req *userv1.SignInReq) (resp *userv1.SignInResp, err error)
	SignOut(ctx context.Context, req *userv1.SignOutReq) (resp *userv1.SignOutResp, err error)
	SignOutByUid(ctx context.Context, req *userv1.SignOutByUidReq) (resp *userv1.SignOutByUidResp, err error)
	Refresh(ctx context.Context, req *userv1.RefreshReq) (resp *userv1.RefreshResp, err error)
	VerifyToken(ctx context.Context, req *userv1.VerifyTokenReq) (resp *userv1.VerifyTokenResp, err error)
	GetActiveSignInLogs(ctx context.Context, req *userv1.GetActiveSignInLogsReq) (resp *userv1.GetActiveSignInLogsResp, err error)
	PagingGetSignInLogs(ctx context.Context, req *userv1.PagingGetSignInLogsReq) (resp *userv1.PagingGetSignInLogsResp, err error)
	AddSessionToBlockList(ctx context.Context, req *userv1.AddSessionToBlockListReq) (resp *userv1.AddSessionToBlockListResp, err error)
}

type Impl struct {
	authHandlers map[enumsv1.AuthType]Authenticator
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
