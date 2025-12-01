package sms

import (
	"context"

	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
)

type ISms interface {
	SendSingleSms(ctx context.Context, req *msgv1.SendSmsReq) (err error)
	QuerySmsStatistics(ctx context.Context, req *msgv1.QuerySmsStatisticsReq) (*msgv1.QuerySmsStatisticsResp, error)
	QuerySmsSendDetail(ctx context.Context, req *msgv1.QuerySmsSendDetailReq) (*msgv1.QuerySmsSendDetailResp, error)
}

func NewSms(cfg *msgv1.SmsConfig) map[enumsv1.MessageSenderVendor]map[string]ISms {
	providers := make(map[enumsv1.MessageSenderVendor]map[string]ISms, len(cfg.Providers))
	for _, provider := range cfg.Providers {
		_, ok := providers[provider.GetVendor()]
		if !ok {
			providers[provider.GetVendor()] = make(map[string]ISms)
		}
		providers[provider.GetVendor()][provider.GetAccount()] = newProvider(provider)
	}
	return providers
}

func newProvider(provider *msgv1.SmsProvider) ISms {
	vendor := provider.GetVendor()
	switch vendor {
	case enumsv1.MessageSenderVendor_MESSAGE_SENDER_VENDOR_ALIYUN:
		return NewAli(provider)
	}
	panic("unsupported vendor type: " + provider.Vendor.String())
}
