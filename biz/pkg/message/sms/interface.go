package sms

import (
	"context"

	"github.com/byteflowing/base/biz/config"
	dalModel "github.com/byteflowing/base/biz/dal/model"
	"github.com/byteflowing/base/biz/pkg/message/model"
	"github.com/byteflowing/base/kitex_gen/base"
)

type SenderApi string

const (
	ApiSendMessage     SenderApi = "SendMessage"
	ApiQuerySendDetail SenderApi = "QuerySendDetail"
)

type Sender interface {
	SendMessage(ctx context.Context, req *model.SendSmsMessageReq) (err error)
	QuerySendDetail(ctx context.Context, message *dalModel.MessageSms) (err error)
	GetProvider() base.SmsProvider
	GetConfig() *config.SmsProvider
	Allow(api SenderApi) bool
	Wait(ctx context.Context, api SenderApi) error
}

type Store interface {
	Save(ctx context.Context, message *dalModel.MessageSms) error
	Update(ctx context.Context, message *dalModel.MessageSms) error
	GetSendingMessages(ctx context.Context, provider base.SmsProvider, limit uint32) ([]*dalModel.MessageSms, error)
}
