package user

import "context"

type EmailPassword struct{}

func (e *EmailPassword) AuthType() Type {
	return AuthTypeEmailPassword
}

func (e *EmailPassword) Authenticate(ctx context.Context, req *LoginReq) (resp *LoginResp, err error) {
	//TODO implement me
	panic("implement me")
}
