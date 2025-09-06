package sms

import (
	"context"
	"errors"

	"github.com/byteflowing/base/ecode"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	smsv1 "github.com/byteflowing/proto/gen/go/sms/v1"
)

type Provider interface {
	SendSms(ctx context.Context, req *smsv1.SendSmsReq) (err error)
}

type Sms interface {
	SendSms(ctx context.Context, req *smsv1.SendSmsReq) (resp *smsv1.SendSmsResp, err error)
}

type Impl struct {
	providers map[enumsv1.SmsVendor]map[string]Provider
}

func New(c *smsv1.Sms) Sms {
	if c == nil {
		return nil
	}
	provider := make(map[enumsv1.SmsVendor]map[string]Provider, len(c.Providers))
	for _, v := range c.Providers {
		_, ok := provider[v.GetVendor()]
		if !ok {
			provider[v.GetVendor()] = make(map[string]Provider)
		}
		provider[v.GetVendor()][v.Account] = newProvider(v)
	}
	return &Impl{
		providers: provider,
	}
}

func (i *Impl) SendSms(ctx context.Context, req *smsv1.SendSmsReq) (resp *smsv1.SendSmsResp, err error) {
	p, err := i.getProvider(req.Vendor, req.Account)
	if err != nil {
		return nil, err
	}
	if req.PhoneNumber == nil || req.PhoneNumber.Number == "" {
		return nil, ecode.ErrPhoneIsEmpty
	}
	err = p.SendSms(ctx, req)
	return nil, err
}

func newProvider(c *smsv1.SmsProvider) Provider {
	vendor := c.GetVendor()
	switch vendor {
	case enumsv1.SmsVendor_SMS_VENDOR_ALIYUN:
		return NewAli(c)
	case enumsv1.SmsVendor_SMS_VENDOR_VOLC:
		return NewVolc(c)
	}
	panic("unsupported vendor type: " + c.Vendor.String())
}

func (i *Impl) getProvider(v enumsv1.SmsVendor, account string) (provider Provider, err error) {
	ps, ok := i.providers[v]
	if !ok {
		return nil, errors.New("provider not exist")
	}
	p, ok := ps[account]
	if !ok {
		return nil, errors.New("account not exist")
	}
	return p, nil
}
