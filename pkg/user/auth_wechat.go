package user

import "context"

type WeChat struct{}

func (w *WeChat) AuthType() Type {
	return AuthTypeWechat
}

func (w *WeChat) Authenticate(ctx context.Context, req *LoginReq) (resp *LoginResp, err error) {
	//TODO implement me
	panic("implement me")
}
