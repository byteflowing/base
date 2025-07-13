package message

import (
	"context"

	"github.com/byteflowing/base/kitex_gen/base"
)

type Message interface {
	SendCaptcha(ctx context.Context, req *base.SendCaptchaReq) (resp *base.SendCaptchaResp, err error)
	VerifyCaptcha(ctx context.Context, req *base.VerifyCaptchaReq) (resp *base.VerifyCaptchaResp, err error)
	PagingGetSmsMessages(ctx context.Context, req *base.PagingGetSmsMessagesReq) (resp *base.PagingGetSmsMessagesResp, err error)
}
