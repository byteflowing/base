package sms

import (
	"context"
	"time"

	"github.com/byteflowing/go-common/3rd/aliyun/sms"
	"github.com/byteflowing/go-common/ratelimit"
	smsv1 "github.com/byteflowing/proto/gen/go/sms/v1"
)

type Ali struct {
	cli         *sms.Sms
	rateLimiter *ratelimit.Limiter
}

func NewAli(c *smsv1.SmsProvider) *Ali {
	cli, err := sms.New(c)
	if err != nil {
		panic(err)
	}
	return &Ali{
		cli:         cli,
		rateLimiter: ratelimit.NewLimiter(1*time.Second, uint64(c.Limit), uint64(c.Limit)),
	}
}

func (a *Ali) SendSms(ctx context.Context, req *smsv1.SendSmsReq) (err error) {
	if err = a.rateLimiter.Wait(ctx); err != nil {
		return
	}
	_, err = a.cli.SendSms(req)
	return err
}
