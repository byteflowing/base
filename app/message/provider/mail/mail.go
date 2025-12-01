package mail

import (
	"context"

	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
)

type IMail interface {
	SendMail(ctx context.Context, req *msgv1.SendMailReq) (err error)
	QueryMailSendStatistics(ctx context.Context, req *msgv1.QueryMailSendStatisticsReq) (resp *msgv1.QueryMailSendStatisticsResp, err error)
	QueryMailSendDetails(ctx context.Context, req *msgv1.QueryMailSendDetailsReq) (resp *msgv1.QueryMailSendDetailsResp, err error)
	QueryMailTracks(ctx context.Context, req *msgv1.QueryMailTracksReq) (resp *msgv1.QueryMailTracksResp, err error)
}

func NewMail(
	cfg *msgv1.MailConfig,
) map[enumsv1.MessageSenderVendor]map[string]IMail {
	providers := make(map[enumsv1.MessageSenderVendor]map[string]IMail)
	for _, provider := range cfg.Providers {
		_, ok := providers[provider.Vendor]
		if !ok {
			providers[provider.Vendor] = make(map[string]IMail)
		}
		providers[provider.GetVendor()][provider.GetAccount()] = newProvider(provider)
	}
	return providers
}

func newProvider(provider *msgv1.MailProvider) IMail {
	switch provider.Vendor {
	case enumsv1.MessageSenderVendor_MESSAGE_SENDER_VENDOR_ALIYUN:
		return NewAli(provider)
	}
	panic("unsupported vendor type: " + provider.Vendor.String())
}
