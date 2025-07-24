package message

import (
	"context"

	"github.com/byteflowing/base/biz/constant"
	"github.com/byteflowing/base/kitex_gen/base"
)

type EmptyImpl struct{}

func (e *EmptyImpl) SendCaptcha(ctx context.Context, req *base.SendCaptchaReq) (resp *base.SendCaptchaResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) VerifyCaptcha(ctx context.Context, req *base.VerifyCaptchaReq) (resp *base.VerifyCaptchaResp, err error) {
	return nil, constant.ErrNotActive
}

func (e *EmptyImpl) PagingGetSmsMessages(ctx context.Context, req *base.PagingGetSmsMessagesReq) (resp *base.PagingGetSmsMessagesResp, err error) {
	return nil, constant.ErrNotActive
}
