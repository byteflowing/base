package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"sync"

	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
	"github.com/wneessen/go-mail"
)

var dialOnce sync.Once

type SMTP struct {
	cli *mail.Client
}

// NewSMTP 创建smtp实例，cron用于检测连接是否断联
func NewSMTP(opts *msgv1.SMTP) (smtp *SMTP, err error) {
	var clientOpts []mail.Option
	clientOpts = append(
		clientOpts,
		mail.WithUsername(opts.UserName),
		mail.WithPassword(opts.Password),
		mail.WithPort(int(opts.Port)),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
	)
	if opts.Timeout.Seconds > 0 {
		clientOpts = append(clientOpts, mail.WithTimeout(opts.Timeout.AsDuration()))
	}
	if opts.SkipTls {
		clientOpts = append(clientOpts, mail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}
	cli, err := mail.NewClient(opts.Host, clientOpts...)
	if err != nil {
		return nil, err
	}
	smtp = &SMTP{
		cli: cli,
	}
	return smtp, nil
}

// DialAndSend 每次都会建立连接发送邮件关闭连接
// 会根据初始化时提供的限流参数进行限流，如果被限流会阻塞，可以通过ctx传入超时控制
func (s *SMTP) DialAndSend(ctx context.Context, mails ...*msgv1.SendMailReq) error {
	if len(mails) == 0 {
		return errors.New("mail list is empty")
	}
	messages, err := s.convert2Messages(mails...)
	if err != nil {
		return err
	}
	return s.cli.DialAndSendWithContext(ctx, messages...)
}

func (s *SMTP) Dial(ctx context.Context) error {
	return s.dial(ctx)
}

// Close 断开与smtp server的连接
func (s *SMTP) Close() error {
	return s.cli.Close()
}

// Reset 重置与smtp server的连接状态
func (s *SMTP) Reset() error {
	return s.cli.Reset()
}

// Send 调用前必须先调用Dial建立连接，关闭时需调用Close关闭连接
func (s *SMTP) Send(ctx context.Context, mails ...*msgv1.SendMailReq) error {
	if len(mails) == 0 {
		return errors.New("mail list is empty")
	}
	if err := s.dial(ctx); err != nil {
		return err
	}
	messages, err := s.convert2Messages(mails...)
	if err != nil {
		return err
	}
	if err = s.cli.Send(messages...); err != nil {
		// 如果发送失败，尝试重拨
		if err := s.cli.DialWithContext(ctx); err != nil {
			return err
		}
		return s.cli.Send(messages...)
	}
	return nil
}

// Dial 与smtp server建立连接
func (s *SMTP) dial(ctx context.Context) (err error) {
	dialOnce.Do(func() {
		err = s.cli.DialWithContext(ctx)
	})
	return err
}

func (s *SMTP) convert2Messages(mails ...*msgv1.SendMailReq) ([]*mail.Msg, error) {
	var messages []*mail.Msg
	for _, m := range mails {
		message := mail.NewMsg()
		if m.From != nil {
			if m.From.Name != nil && *m.From.Name != "" {
				if err := message.FromFormat(*m.From.Name, m.From.Address); err != nil {
					return nil, err
				}
			} else {
				if err := message.From(m.From.Address); err != nil {
					return nil, err
				}
			}
		}
		for _, to := range m.To {
			if to.Name != nil && *to.Name != "" {
				if err := message.AddToFormat(*to.Name, to.Address); err != nil {
					return nil, err
				}
			} else {
				if err := message.AddTo(to.Address); err != nil {
					return nil, err
				}
			}
		}
		for _, cc := range m.Cc {
			if cc.Name != nil && *cc.Name != "" {
				if err := message.AddCcFormat(*cc.Name, cc.Address); err != nil {
					return nil, err
				}
			} else {
				if err := message.AddCc(cc.Address); err != nil {
					return nil, err
				}
			}
		}
		for _, bcc := range m.Bcc {
			if bcc.Name != nil && *bcc.Name != "" {
				if err := message.AddBccFormat(*bcc.Name, bcc.Address); err != nil {
					return nil, err
				}
			} else {
				if err := message.AddBcc(bcc.Address); err != nil {
					return nil, err
				}
			}
		}
		for _, attachment := range m.Attachments {
			message.AttachFile(attachment)
		}
		message.Subject(m.Subject)
		message.SetBodyString(toRealType(m.ContentType), m.Content)
		messages = append(messages, message)
	}
	return messages, nil
}

func toRealType(t enumv1.MailContentType) mail.ContentType {
	switch t {
	case enumv1.MailContentType_MAIL_CONTENT_TYPE_HTML:
		return mail.TypeTextHTML
	case enumv1.MailContentType_MAIL_CONTENT_TYPE_TEXT:
		return mail.TypeTextPlain
	default:
		return mail.TypeTextHTML
	}
}
