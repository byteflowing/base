package captcha

import (
	"context"

	messageV1 "github.com/byteflowing/base/gen/message/v1"
)

type Captcha interface {
	Send(ctx context.Context, req *messageV1.SendCaptchaReq) (resp *messageV1.SendCaptchaResp, err error)
	Verify(ctx context.Context, req *messageV1.VerifyCaptchaReq) (resp *messageV1.VerifyCaptchaResp, err error)
}
