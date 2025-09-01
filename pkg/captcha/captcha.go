package captcha

import (
	"context"
	"errors"

	configv1 "github.com/byteflowing/base/gen/config/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	messageV1 "github.com/byteflowing/base/gen/msg/v1"
	"github.com/byteflowing/base/pkg/msg/mail"
	"github.com/byteflowing/base/pkg/msg/sms"
	"github.com/byteflowing/go-common/redis"
)

type Provider interface {
	Send(ctx context.Context, req *messageV1.SendCaptchaReq) (resp *messageV1.SendCaptchaResp, err error)
	Verify(ctx context.Context, req *messageV1.VerifyCaptchaReq) (resp *messageV1.VerifyCaptchaResp, err error)
}

type Captcha interface {
	SendCaptcha(ctx context.Context, req *messageV1.SendCaptchaReq) (resp *messageV1.SendCaptchaResp, err error)
	VerifyCaptcha(ctx context.Context, req *messageV1.VerifyCaptchaReq) (resp *messageV1.VerifyCaptchaResp, err error)
}

func NewCaptcha(
	c *configv1.Captcha,
	rdb *redis.Redis,
	sms sms.Sms,
	mail mail.Mail,
) Captcha {
	if c == nil {
		return nil
	}
	if rdb == nil {
		panic("rdb is nil")
	}
	providers := make(map[enumsv1.MessageSenderType]Provider, len(c.Providers))
	for _, p := range c.Providers {
		switch p.Sender {
		case enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS:
			if sms == nil {
				panic("sms is nil")
			}
			providers[p.Sender] = NewSmsCaptcha(rdb, sms, p)
		case enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL:
			if mail == nil {
				panic("mail is nil")
			}
			providers[p.Sender] = NewMailCaptcha(rdb, mail, p)
		}
	}
	return &Impl{providers: providers}
}

type Impl struct {
	providers map[enumsv1.MessageSenderType]Provider
}

func (i *Impl) SendCaptcha(ctx context.Context, req *messageV1.SendCaptchaReq) (resp *messageV1.SendCaptchaResp, err error) {
	sender, err := i.getSender(req.SenderType)
	if err != nil {
		return nil, err
	}
	return sender.Send(ctx, req)
}

func (i *Impl) VerifyCaptcha(ctx context.Context, req *messageV1.VerifyCaptchaReq) (resp *messageV1.VerifyCaptchaResp, err error) {
	sender, err := i.getSender(req.SenderType)
	if err != nil {
		return nil, err
	}
	return sender.Verify(ctx, req)
}

func (i *Impl) getSender(senderType enumsv1.MessageSenderType) (sender Provider, err error) {
	sender, ok := i.providers[senderType]
	if !ok {
		return nil, errors.New("sender not exist")
	}
	return sender, nil
}
