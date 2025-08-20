package sms

import (
	"context"
	"fmt"
	"time"

	configv1 "github.com/byteflowing/base/gen/config/v1"
	messagev1 "github.com/byteflowing/base/gen/message/v1"
	"github.com/byteflowing/go-common/3rd/aliyun/sms"
	"github.com/byteflowing/go-common/ratelimit"
)

type Ali struct {
	cli         *sms.Sms
	rateLimiter *ratelimit.Limiter
}

func NewAli(c *configv1.SmsProvider) *Ali {
	cli, err := sms.New(&sms.Opts{
		AccessKeyId:     c.AccessKey,
		AccessKeySecret: c.AccessKey,
		SecurityToken:   c.SecurityToken,
	})
	if err != nil {
		panic(err)
	}
	return &Ali{
		cli:         cli,
		rateLimiter: ratelimit.NewLimiter(1*time.Second, uint64(c.Limit), uint64(c.Limit)),
	}
}

func (a *Ali) SendSms(ctx context.Context, req *messagev1.SendSmsReq) (err error) {
	if err = a.rateLimiter.Wait(ctx); err != nil {
		return
	}
	resp, err := a.cli.SendSms(&sms.SendSmsReq{
		PhoneNumbers:  req.PhoneNumber.Number,
		SignName:      req.SignName,
		TemplateCode:  req.TemplateCode,
		TemplateParam: req.TemplateParams,
	})
	if err != nil {
		return
	}
	return a.parseErr(resp.Common)
}

func (a *Ali) parseErr(commonResp *sms.CommonResp) error {
	if commonResp.Code != "OK" {
		return fmt.Errorf("code: %s, msg: %s, biz_id: %s, request_id: %s",
			commonResp.Code,
			commonResp.Message,
			commonResp.BizId,
			commonResp.RequestId)
	}
	return nil
}
