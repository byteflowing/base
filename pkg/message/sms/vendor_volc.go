package sms

import (
	"context"
	"github.com/byteflowing/go-common/3rd/bytedance/volc"
	"github.com/byteflowing/go-common/ratelimit"
	"time"
)

type Volc struct {
	config      *ProviderConfig
	cli         *volc.Sms
	rateLimiter *ratelimit.Limiter
}

func NewVolc(c *ProviderConfig) *Volc {
	cli := volc.NewSms(&volc.SmsOpts{
		AccessKeyId:     c.AccessKey,
		AccessKeySecret: c.SecretKey,
	})
	return &Volc{
		config:      c,
		cli:         cli,
		rateLimiter: ratelimit.NewLimiter(1*time.Second, uint64(c.SendMessageQPS), uint64(c.SendMessageQPS)),
	}
}

func (v *Volc) SendMessage(ctx context.Context, req *SendMessageReq) (err error) {
	if err = v.rateLimiter.Wait(ctx); err != nil {
		return
	}
	_, err = v.cli.SendSms(&volc.SendSmsReq{
		SmsAccount:    v.config.Account,
		Sign:          req.SignName,
		TemplateID:    req.TemplateCode,
		TemplateParam: req.TemplateParam,
		PhoneNumber:   req.PhoneNumber,
	})
	return
}
