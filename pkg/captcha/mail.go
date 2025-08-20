package captcha

import (
	"context"
	"errors"

	"github.com/byteflowing/base/ecode"
	configv1 "github.com/byteflowing/base/gen/config/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	messageV1 "github.com/byteflowing/base/gen/message/v1"
	"github.com/byteflowing/base/pkg/message/mail"
	"github.com/byteflowing/go-common/redis"
)

type MailCaptcha struct {
	mail    mail.Mail
	captcha *captcha
}

func NewMailCaptcha(rdb *redis.Redis, mail mail.Mail, c *configv1.Captcha) *MailCaptcha {
	return &MailCaptcha{
		mail:    mail,
		captcha: newCaptcha(rdb, c),
	}
}

func (m *MailCaptcha) Send(ctx context.Context, req *messageV1.SendCaptchaReq) (resp *messageV1.SendCaptchaResp, err error) {
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
	resp = &messageV1.SendCaptchaResp{
		Data: &messageV1.SendCaptchaResp_Data{
			Token: token,
			Limit: limit,
		},
	}
	return resp, nil
}

func (m *MailCaptcha) Verify(ctx context.Context, req *messageV1.VerifyCaptchaReq) (resp *messageV1.VerifyCaptchaResp, err error) {
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
