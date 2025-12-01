package sms

import (
	"fmt"

	"github.com/byteflowing/go-common/jsonx"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
	"github.com/volcengine/volc-sdk-golang/service/sms"
)

// 文档：https://www.volcengine.com/docs/6361/67380
// go sdk：https://www.volcengine.com/docs/6361/1109261

type Sms struct {
	accessKeyId     string
	accessKeySecret string
	cli             *sms.SMS
}

func NewSms(opts *msgv1.SmsProvider) *Sms {
	cli := sms.DefaultInstance
	cli.Client.SetAccessKey(opts.AccessKey)
	cli.Client.SetSecretKey(opts.SecretKey)
	return &Sms{
		accessKeyId:     opts.AccessKey,
		accessKeySecret: opts.SecretKey,
		cli:             cli,
	}
}

func (s *Sms) SendSms(req *msgv1.SendSmsReq) (resp *msgv1.SendSmsResp, err error) {
	params, err := jsonx.MarshalToString(req.TemplateParams)
	if err != nil {
		return nil, err
	}
	res, _, err := s.cli.Send(&sms.SmsRequest{
		SmsAccount:    req.Account,
		Sign:          req.SignName,
		TemplateID:    req.TemplateCode,
		TemplateParam: params,
		PhoneNumbers:  req.PhoneNumber.Number,
	})
	err = s.parseErr(res, err)
	return
}

func (s *Sms) parseErr(resp *sms.SmsResponse, err error) error {
	if err != nil {
		if resp != nil {
			return fmt.Errorf(
				"requestID: %s, action: %s, version: %s, service: %s, grpc_geo: %s, messageID: %v, err: %v",
				resp.ResponseMetadata.RequestId,
				resp.ResponseMetadata.Action,
				resp.ResponseMetadata.Version,
				resp.ResponseMetadata.Service,
				resp.ResponseMetadata.Region,
				resp.Result.MessageID,
				err,
			)
		}
		return err
	}
	return nil
}
