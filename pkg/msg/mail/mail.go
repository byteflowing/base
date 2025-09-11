package mail

import (
	"context"
	"errors"

	"github.com/byteflowing/go-common/mail"
	"github.com/byteflowing/go-common/ratelimit"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mailv1 "github.com/byteflowing/proto/gen/go/mail/v1"
)

type Mail interface {
	SendMail(ctx context.Context, req *mailv1.SendMailReq) (resp *mailv1.SendMailResp, err error)
	GetSMTP(vendor enumsv1.MailVendor, account string) (smtp *mail.SMTP, limiter *ratelimit.Limiter, err error)
}

type Impl struct {
	clients map[enumsv1.MailVendor]map[string]*mail.SMTP
	limits  map[enumsv1.MailVendor]map[string]*ratelimit.Limiter
}

func New(c *mailv1.MailConfig) Mail {
	clients := make(map[enumsv1.MailVendor]map[string]*mail.SMTP, len(c.Providers))
	limits := make(map[enumsv1.MailVendor]map[string]*ratelimit.Limiter)
	for _, provider := range c.Providers {
		vendor := provider.GetVendor()
		_, ok := clients[vendor]
		if !ok {
			clients[vendor] = make(map[string]*mail.SMTP)
		}
		m, err := mail.NewSMTP(provider.Smtp)
		if err != nil {
			panic(err)
		}
		clients[vendor][provider.Account] = m
		limits[vendor][provider.Account] = ratelimit.NewLimiter(provider.LimitDuration.AsDuration(), uint64(provider.LimitMax), uint64(provider.LimitMax))
	}
	return &Impl{
		clients: clients,
		limits:  limits,
	}
}

// SendMail 每次发送都会重新建立smtp连接，发送完断开连接
// 如果有连续批量邮件发送任务，使用GetSMTP获取smtp实力，使用Dial建立连接，使用Send发送邮件，使用Close关闭连接
func (i *Impl) SendMail(ctx context.Context, req *mailv1.SendMailReq) (resp *mailv1.SendMailResp, err error) {
	limiter, err := i.getLimiter(req.Vendor, req.Account)
	if err != nil {
		return nil, err
	}
	if err = limiter.Wait(ctx); err != nil {
		return nil, err
	}
	smtp, err := i.getSMTP(req.Vendor, req.Account)
	if err != nil {
		return nil, err
	}
	err = smtp.DialAndSend(ctx, req)
	return nil, err
}

func (i *Impl) GetSMTP(vendor enumsv1.MailVendor, account string) (smtp *mail.SMTP, limiter *ratelimit.Limiter, err error) {
	smtp, err = i.getSMTP(vendor, account)
	if err != nil {
		return nil, nil, err
	}
	limiter, err = i.getLimiter(vendor, account)
	if err != nil {
		return nil, nil, err
	}
	return smtp, limiter, nil
}

func (i *Impl) getSMTP(vendor enumsv1.MailVendor, account string) (*mail.SMTP, error) {
	p, ok := i.clients[vendor]
	if !ok {
		return nil, errors.New("vendor not exist")
	}
	smtp, ok := p[account]
	if !ok {
		return nil, errors.New("account not exist")
	}
	return smtp, nil
}

func (i *Impl) getLimiter(vendor enumsv1.MailVendor, account string) (*ratelimit.Limiter, error) {
	l, ok := i.limits[vendor]
	if !ok {
		return nil, errors.New("vendor not exist")
	}
	limiter, ok := l[account]
	if !ok {
		return nil, errors.New("account not exist")
	}
	return limiter, nil
}
