package user

import (
	"context"
)

type UserNamePassword struct {
}

func (u *UserNamePassword) AuthType() Type {
	return AuthTypeUserNamePassword
}

func (u *UserNamePassword) Authenticate(ctx context.Context, req *LoginReq) (resp *LoginResp, err error) {
	//TODO implement me
	panic("implement me")
}
