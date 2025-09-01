package captcha

import (
	"context"
	"errors"
	"fmt"

	"github.com/byteflowing/base/ecode"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	configv1 "github.com/byteflowing/base/gen/config/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	messageV1 "github.com/byteflowing/base/gen/msg/v1"
	"github.com/byteflowing/base/pkg/msg/sms"
	"github.com/byteflowing/go-common/redis"
)

type SmsCaptcha struct {
	sms     sms.Sms
	captcha *captcha
}

func NewSmsCaptcha(rdb *redis.Redis, sms sms.Sms, c *configv1.CaptchaProvider) *SmsCaptcha {
	return &SmsCaptcha{
		sms:     sms,
		captcha: newCaptcha(rdb, c),
	}
}

func (s *SmsCaptcha) Send(ctx context.Context, req *messageV1.SendCaptchaReq) (resp *messageV1.SendCaptchaResp, err error) {
	if req.SenderType != enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS {
		return nil, errors.New("sender type must be SMS")
	}
	smsReq := req.GetSms()
	if smsReq == nil {
		return nil, errors.New("sms request is nil")
	}
	target := getPhoneTarget(smsReq.PhoneNumber)
	token, limit, err := s.captcha.send(ctx, target, req.Captcha, func() error {
		_, err = s.sms.SendSms(ctx, smsReq)
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

func (s *SmsCaptcha) Verify(ctx context.Context, req *messageV1.VerifyCaptchaReq) (resp *messageV1.VerifyCaptchaResp, err error) {
	if req.SenderType != enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS {
		return nil, errors.New("sender type must be SMS")
	}
	target := getPhoneTarget(req.GetPhoneNumber())
	ok, err := s.captcha.verify(ctx, target, req.Token, req.Captcha, enumsv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ecode.ErrCaptchaMisMatch
	}
	return nil, nil
}

func getPhoneTarget(phone *commonv1.PhoneNumber) string {
	return fmt.Sprintf("%s%s", phone.GetCountryCode(), phone.GetNumber())
}
