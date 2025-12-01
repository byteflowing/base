package sms

import (
	"context"

	"github.com/byteflowing/base/pkg/sdk/alibaba/aliyun/sms"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
)

type Ali struct {
	cli *sms.Sms
}

func NewAli(c *msgv1.SmsProvider) *Ali {
	cli, err := sms.New(c)
	if err != nil {
		panic(err)
	}
	return &Ali{
		cli: cli,
	}
}

func (a *Ali) SendSingleSms(_ context.Context, req *msgv1.SendSmsReq) (err error) {
	_, err = a.cli.SendSms(req)
	return err
}

func (a *Ali) QuerySmsStatistics(ctx context.Context, req *msgv1.QuerySmsStatisticsReq) (*msgv1.QuerySmsStatisticsResp, error) {
	return a.cli.QuerySendStatistics(req)
}

func (a *Ali) QuerySmsSendDetail(ctx context.Context, req *msgv1.QuerySmsSendDetailReq) (*msgv1.QuerySmsSendDetailResp, error) {
	return a.cli.QuerySendDetail(req)
}
