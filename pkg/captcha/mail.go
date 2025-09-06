package captcha

import (
	"context"
	"errors"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/msg/mail"
	"github.com/byteflowing/go-common/redis"
	captchav1 "github.com/byteflowing/proto/gen/go/captcha/v1"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
)

type MailCaptcha struct {
	mail    mail.Mail
	captcha *captcha
}

func NewMailCaptcha(rdb *redis.Redis, mail mail.Mail, c *captchav1.CaptchaProvider) *MailCaptcha {
	return &MailCaptcha{
		mail:    mail,
		captcha: newCaptcha(rdb, c),
	}
}

func (m *MailCaptcha) Send(ctx context.Context, req *captchav1.SendCaptchaReq) (resp *captchav1.SendCaptchaResp, err error) {
	if req.SenderType != enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL {
		return nil, errors.New("sender type must be MAIL")
	}
	mailReq := req.GetMail()
	if mailReq == nil {
		return nil, errors.New("mail request is nil")
	}
	if len(mailReq.To) != 1 {
		return nil, errors.New("send captcha mail address to must be 1")
	}
	token, limit, err := m.captcha.send(ctx, mailReq.To[0].Address, req.Captcha, func() error {
		_, err = m.mail.SendMail(ctx, mailReq)
		return err
	})
	if err != nil {
		return nil, err
	}
	resp = &captchav1.SendCaptchaResp{
		Data: &captchav1.SendCaptchaResp_Data{
			Token: token,
			Limit: limit,
		},
	}
	return resp, nil
}

func (m *MailCaptcha) Verify(ctx context.Context, req *captchav1.VerifyCaptchaReq) (resp *captchav1.VerifyCaptchaResp, err error) {
	if req.SenderType != enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL {
		return nil, errors.New("sender type must be MAIL")
	}
	mailAddr := req.GetEmail()
	if mailAddr == "" {
		return nil, errors.New("email address is empty")
	}
	ok, err := m.captcha.verify(ctx, mailAddr, req.Token, req.Captcha, enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ecode.ErrCaptchaMisMatch
	}
	return nil, nil
}
