package sms

import (
	"context"
	"time"

	configv1 "github.com/byteflowing/base/gen/config/v1"
	messagev1 "github.com/byteflowing/base/gen/message/v1"
	"github.com/byteflowing/go-common/3rd/bytedance/volc"
	"github.com/byteflowing/go-common/ratelimit"
)

type Volc struct {
	config      *configv1.SmsProvider
	cli         *volc.Sms
	rateLimiter *ratelimit.Limiter
}

func NewVolc(c *configv1.SmsProvider) *Volc {
	cli := volc.NewSms(&volc.SmsOpts{
		AccessKeyId:     c.AccessKey,
		AccessKeySecret: c.SecretKey,
	})
	return &Volc{
		config:      c,
		cli:         cli,
		rateLimiter: ratelimit.NewLimiter(1*time.Second, uint64(c.Limit), uint64(c.Limit)),
	}
}

func (v *Volc) SendSms(ctx context.Context, req *messagev1.SendSmsReq) (err error) {
	if err = v.rateLimiter.Wait(ctx); err != nil {
		return
	}
	_, err = v.cli.SendSms(&volc.SendSmsReq{
		SmsAccount:    v.config.Account,
		Sign:          req.SignName,
		TemplateID:    req.TemplateCode,
		TemplateParam: req.TemplateParams,
		PhoneNumber:   req.PhoneNumber.Number,
	})
	return
}
