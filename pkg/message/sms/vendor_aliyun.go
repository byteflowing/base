package sms

import (
	"context"
	"fmt"
	"time"

	"github.com/byteflowing/go-common/3rd/aliyun/sms"
	"github.com/byteflowing/go-common/ratelimit"
)

type Ali struct {
	cli         *sms.Sms
	rateLimiter *ratelimit.Limiter
}

func NewAli(c *ProviderConfig) *Ali {
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
		rateLimiter: ratelimit.NewLimiter(1*time.Second, uint64(c.SendMessageQPS), uint64(c.SendMessageQPS)),
	}
}

func (a *Ali) SendMessage(ctx context.Context, req *SendMessageReq) (err error) {
	if err = a.rateLimiter.Wait(ctx); err != nil {
		return
	}
	resp, err := a.cli.SendSms(&sms.SendSmsReq{
		PhoneNumbers:  req.PhoneNumber,
		SignName:      req.SignName,
		TemplateCode:  req.TemplateCode,
		TemplateParam: req.TemplateParam,
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
