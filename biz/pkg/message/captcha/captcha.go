package captcha

import (
	"context"

	"github.com/byteflowing/base/kitex_gen/base"
)

type Captcha interface {
	Send(ctx context.Context, req *base.SendCaptchaReq) (resp *base.SendCaptchaResp, err error)
	Verify(ctx context.Context, req *base.VerifyCaptchaReq) (bool, error)
}
