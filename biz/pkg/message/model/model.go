package model

import "github.com/byteflowing/base/kitex_gen/base"

type SmsMessage struct {
	Type           base.MessageType
	Provider       base.SmsProvider
	CaptchaType    *base.CaptchaType
	SenderID       int64
	PhoneNumber    string
	SignName       string
	TemplateCode   string
	TemplateParams map[string]string
}

type SendSmsMessageReq struct {
	*SmsMessage
}

type QuerySmsSendDetailReq struct {
	Phone *string
	BizId *string
	Date  *string
}

type QuerySmsSendDetailResp struct {
	ErrCode     *string
	ReceiveDate *string
	SendDate    *string
	Content     *string
	Status      int16
}
