package mail

import (
	"context"

	"github.com/byteflowing/base/pkg/mail"
	aliMail "github.com/byteflowing/base/pkg/sdk/alibaba/aliyun/mail"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
)

type Ali struct {
	smtpCli *mail.SMTP
	apiCli  *aliMail.Mail
}

func NewAli(provider *msgv1.MailProvider) *Ali {
	smtpCli, err := mail.NewSMTP(provider.Smtp)
	if err != nil {
		panic(err)
	}
	apiCli, err := aliMail.NewMail(provider.ApiConfig)
	if err != nil {
		panic(err)
	}
	return &Ali{
		smtpCli: smtpCli,
		apiCli:  apiCli,
	}
}

func (a *Ali) SendMail(ctx context.Context, req *msgv1.SendMailReq) (err error) {
	return a.smtpCli.Send(ctx, req)
}

func (a *Ali) QueryMailSendStatistics(_ context.Context, req *msgv1.QueryMailSendStatisticsReq) (resp *msgv1.QueryMailSendStatisticsResp, err error) {
	return a.apiCli.QueryMailSendStatistics(req)
}

func (a *Ali) QueryMailSendDetails(_ context.Context, req *msgv1.QueryMailSendDetailsReq) (resp *msgv1.QueryMailSendDetailsResp, err error) {
	return a.apiCli.QueryMailSendDetail(req)
}

func (a *Ali) QueryMailTracks(_ context.Context, req *msgv1.QueryMailTracksReq) (resp *msgv1.QueryMailTracksResp, err error) {
	return a.apiCli.QueryMailTracks(req)
}
