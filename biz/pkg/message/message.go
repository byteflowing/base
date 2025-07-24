package message

import (
	"context"

	"github.com/byteflowing/base/biz/config"
	"github.com/byteflowing/base/biz/constant"
	"github.com/byteflowing/base/biz/dal/query"
	"github.com/byteflowing/base/biz/pkg/message/captcha"
	"github.com/byteflowing/base/biz/pkg/message/sms"
	"github.com/byteflowing/base/kitex_gen/base"
	"github.com/byteflowing/go-common/redis"
)

type Message interface {
	SendCaptcha(ctx context.Context, req *base.SendCaptchaReq) (resp *base.SendCaptchaResp, err error)
	VerifyCaptcha(ctx context.Context, req *base.VerifyCaptchaReq) (resp *base.VerifyCaptchaResp, err error)
	PagingGetSmsMessages(ctx context.Context, req *base.PagingGetSmsMessagesReq) (resp *base.PagingGetSmsMessagesResp, err error)
}

type Opts struct {
	Config *config.MessageConfig
	RDB    *redis.Redis
	DB     *query.Query
}

func New(opts *Opts) Message {
	if opts.Config == nil || !opts.Config.Active {
		return new(EmptyImpl)
	}
	cs := make(map[base.MessageSender]captcha.Captcha)
	captchaStore := captcha.NewRedisStore(opts.RDB)
	if opts.Config.Sms != nil {
		smsManager := sms.NewSmsManager(&sms.MangerOpts{
			SmsConfig: opts.Config.Sms,
			RDB:       opts.RDB,
			DB:        opts.DB,
		})
		cs[base.MessageSender_MESSAGE_SENDER_SMS] = captcha.NewSmsCaptcha(&captcha.SmsCaptchaOption{
			SmsConfig:    opts.Config.Sms,
			SmsManager:   smsManager,
			CaptchaStore: captchaStore,
			RDB:          opts.RDB,
		})
	}
	return &Impl{
		captcha: cs,
	}
}

type Impl struct {
	captcha map[base.MessageSender]captcha.Captcha
}

func (i *Impl) getCaptcha(s base.MessageSender) (captcha.Captcha, error) {
	c, ok := i.captcha[s]
	if !ok {
		return nil, constant.ErrNotActive
	}
	return c, nil
}

func (i *Impl) SendCaptcha(ctx context.Context, req *base.SendCaptchaReq) (resp *base.SendCaptchaResp, err error) {
	c, err := i.getCaptcha(req.Sender)
	if err != nil {
		return nil, err
	}
	return c.Send(ctx, req)
}

func (i *Impl) VerifyCaptcha(ctx context.Context, req *base.VerifyCaptchaReq) (resp *base.VerifyCaptchaResp, err error) {
	c, err := i.getCaptcha(req.Sender)
	if err != nil {
		return nil, err
	}
	ok, err := c.Verify(ctx, req)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, constant.ErrCaptchaMisMatch
	}
	return nil, nil
}

func (i *Impl) PagingGetSmsMessages(ctx context.Context, req *base.PagingGetSmsMessagesReq) (resp *base.PagingGetSmsMessagesResp, err error) {
	//TODO implement me
	panic("implement me")
}
