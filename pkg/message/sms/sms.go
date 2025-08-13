package sms

import (
	"context"
	"fmt"

	"github.com/byteflowing/base/pkg/message/captcha"
	"github.com/byteflowing/go-common/redis"
)

type Provider interface {
	SendMessage(ctx context.Context, req *SendMessageReq) (err error)
}

type Sms interface {
	SendCaptcha(ctx context.Context, req *SendCaptchaReq) (token string, limit *captcha.LimitRule, err error)
	VerifyCaptcha(ctx context.Context, req *VerifyCaptchaReq) (ok bool, err error)
}

type Impl struct {
	providers map[Vendor]Provider
	captcha   captcha.Captcha
}

func New(rdb *redis.Redis, c *Config) *Impl {
	provider := make(map[Vendor]Provider, len(c.Providers))
	for _, v := range c.Providers {
		provider[v.GetVendor()] = newProvider(v)
	}
	_captcha := captcha.New(rdb, c.Captcha)
	return &Impl{
		providers: provider,
		captcha:   _captcha,
	}
}

func (i *Impl) SendCaptcha(ctx context.Context, req *SendCaptchaReq) (token string, rule *captcha.LimitRule, err error) {
	return i.captcha.Save(ctx, req.PhoneNumber, req.Captcha, func() error {
		p, err := i.getProvider(req.Vendor)
		if err != nil {
			return err
		}
		return p.SendMessage(ctx, &SendMessageReq{
			PhoneNumber:   req.PhoneNumber,
			SignName:      req.SignName,
			TemplateCode:  req.TemplateCode,
			TemplateParam: req.TemplateParam,
		})
	})
}

func (i *Impl) VerifyCaptcha(ctx context.Context, req *VerifyCaptchaReq) (ok bool, err error) {
	return i.captcha.Verify(ctx, req.Token, req.Captcha)
}

func newProvider(c *ProviderConfig) Provider {
	vendor := c.GetVendor()
	switch vendor {
	case VendorAli:
		return NewAli(c)
	case VendorVolc:
		return NewVolc(c)
	}
	panic("unsupported vendor type:" + vendor)
}

func (i *Impl) getProvider(v Vendor) (p Provider, err error) {
	p, ok := i.providers[v]
	if !ok {
		return nil, fmt.Errorf("unsupported vendor type: %s", v)
	}
	return p, nil
}
