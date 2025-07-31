package user

import "context"

type PhoneCaptcha struct{}

func (p *PhoneCaptcha) AuthType() Type {
	return AuthTypePhoneCaptcha
}

func (p *PhoneCaptcha) Authenticate(ctx context.Context, req *LoginReq) (resp *LoginResp, err error) {
	//TODO implement me
	panic("implement me")
}
