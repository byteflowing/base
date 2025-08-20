package mail

import (
	"context"
	"errors"

	configv1 "github.com/byteflowing/base/gen/config/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	messagev1 "github.com/byteflowing/base/gen/message/v1"
	"github.com/byteflowing/go-common/mail"
	"github.com/byteflowing/go-common/ratelimit"
)

type Mail interface {
	SendMail(ctx context.Context, req *messagev1.SendMailReq) (resp *messagev1.SendMailResp, err error)
	GetSMTP(vendor enumsv1.MailVendor, account string) (smtp *mail.SMTP, limiter *ratelimit.Limiter, err error)
}

type Impl struct {
	clients map[enumsv1.MailVendor]map[string]*mail.SMTP
	limits  map[enumsv1.MailVendor]map[string]*ratelimit.Limiter
}

func New(c *configv1.Mail) *Impl {
	clients := make(map[enumsv1.MailVendor]map[string]*mail.SMTP, len(c.Providers))
	limits := make(map[enumsv1.MailVendor]map[string]*ratelimit.Limiter)
	for _, provider := range c.Providers {
		vendor := provider.GetVendor()
		_, ok := clients[vendor]
		if !ok {
			clients[vendor] = make(map[string]*mail.SMTP)
		}
		m, err := mail.NewSMTP(&mail.SMTPOpts{
			Host:     provider.Smtp.Host,
			Port:     int(provider.Smtp.Port),
			Username: provider.Smtp.UserName,
			Password: provider.Smtp.Password,
			SkipTLS:  provider.Smtp.SkipTls,
			Timeout:  int(provider.Smtp.Timeout.Seconds),
		})
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
func (i *Impl) SendMail(ctx context.Context, req *messagev1.SendMailReq) (resp *messagev1.SendMailResp, err error) {
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
	err = smtp.DialAndSend(ctx, &mail.Mail{
		From:        i.convertAddr(req.From),
		To:          i.convertAddresses(req.To),
		Cc:          i.convertAddresses(req.Cc),
		Bcc:         i.convertAddresses(req.Bcc),
		Subject:     req.Subject,
		ContentType: i.convertContentType(req.ContentType),
		Content:     req.Content,
		Attachments: req.Attachments,
	})
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

func (i *Impl) convertAddr(from *messagev1.MailAddress) *mail.Address {
	return &mail.Address{
		Name: from.GetName(),
		Addr: from.GetAddress(),
	}
}

func (i *Impl) convertAddresses(addresses []*messagev1.MailAddress) []*mail.Address {
	if len(addresses) == 0 {
		return nil
	}
	var adders []*mail.Address
	for _, addr := range addresses {
		adders = append(adders, i.convertAddr(addr))
	}
	return adders
}

func (i *Impl) convertContentType(t enumsv1.MailContentType) mail.ContentType {
	switch t {
	case enumsv1.MailContentType_MAIL_CONTENT_TYPE_TEXT:
		return mail.ContentTypeText
	case enumsv1.MailContentType_MAIL_CONTENT_TYPE_HTML:
		return mail.ContentTypeHTML
	}
	return mail.ContentTypeHTML
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
