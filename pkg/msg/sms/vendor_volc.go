package sms

import (
	"context"
	"time"

	"github.com/byteflowing/go-common/3rd/bytedance/volc"
	"github.com/byteflowing/go-common/ratelimit"
	smsv1 "github.com/byteflowing/proto/gen/go/sms/v1"
)

type Volc struct {
	config      *smsv1.SmsProvider
	cli         *volc.Sms
	rateLimiter *ratelimit.Limiter
}

func NewVolc(c *smsv1.SmsProvider) *Volc {
	cli := volc.NewSms(c)
	return &Volc{
		config:      c,
		cli:         cli,
		rateLimiter: ratelimit.NewLimiter(1*time.Second, uint64(c.Limit), uint64(c.Limit)),
	}
}

func (v *Volc) SendSms(ctx context.Context, req *smsv1.SendSmsReq) (err error) {
	if err = v.rateLimiter.Wait(ctx); err != nil {
		return
	}
	_, err = v.cli.SendSms(req)
	return
}
