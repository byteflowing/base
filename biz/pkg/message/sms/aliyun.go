package sms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/byteflowing/base/biz/config"
	dalModel "github.com/byteflowing/base/biz/dal/model"
	"github.com/byteflowing/base/biz/pkg/message/model"
	"github.com/byteflowing/base/kitex_gen/base"
	"github.com/byteflowing/go-common/3rd/aliyun/sms"
	"github.com/byteflowing/go-common/ratelimit"
	"github.com/byteflowing/go-common/service"
	"github.com/cloudwego/kitex/pkg/klog"
)

type AliSmsSender struct {
	cli       *sms.Sms
	conf      *config.SmsProvider
	store     Store
	limiters  map[SenderApi]*ratelimit.Limiter
	saveTasks *service.TaskRunner
}

func NewAliSmsSender(conf *config.SmsProvider, store Store) *AliSmsSender {
	limiters := make(map[SenderApi]*ratelimit.Limiter, 2)
	limiters[ApiSendMessage] = ratelimit.NewLimiter(1*time.Second, uint64(conf.SendMessageLimit), uint64(conf.SendMessageLimit))
	limiters[ApiQuerySendDetail] = ratelimit.NewLimiter(1*time.Second, uint64(conf.QuerySendDetailLimit), uint64(conf.QuerySendDetailLimit))
	cli, err := sms.New(&sms.Opts{
		AccessKeyId:     conf.AccessKey,
		AccessKeySecret: conf.SecretKey,
		SecurityToken:   conf.SecurityKey,
	})
	if err != nil {
		panic(err)
	}
	return &AliSmsSender{
		cli:       cli,
		conf:      conf,
		store:     store,
		limiters:  limiters,
		saveTasks: service.NewTaskRunner(int(conf.SendMessageLimit * 2)),
	}
}

func (ali *AliSmsSender) SendMessage(ctx context.Context, req *model.SendSmsMessageReq) (err error) {
	r, err := ali.cli.SendSms(&sms.SendSmsReq{
		PhoneNumbers:  req.PhoneNumber,
		SignName:      req.SignName,
		TemplateCode:  req.TemplateCode,
		TemplateParam: req.TemplateParams,
	})
	if err != nil {
		return err
	}
	if r.Common.Code != "OK" {
		return fmt.Errorf("send sms failed, code: %s, msg: %s", r.Common.Code, r.Common.Message)
	}
	ali.saveTasks.Schedule(func() {
		if err := ali.store.Save(ctx, &dalModel.MessageSms{
			MsgType:     int16(req.Type),
			MsgStatus:   int16(base.MessageStatus_MESSAGE_STATUS_SENDING),
			CaptchaType: getCaptchaType(req.CaptchaType),
			Provider:    int32(req.Provider),
			Template:    req.TemplateCode,
			Sign:        req.SignName,
			RequestID:   &r.Common.RequestId,
			BizID:       &r.Common.BizId,
			Phone:       req.PhoneNumber,
			SenderID:    req.SenderID,
			Params:      marshalParams(ctx, req.TemplateParams),
		}); err != nil {
			klog.Errorf("save message to store error: %v", err)
		}
	})
	return nil
}

func (ali *AliSmsSender) QuerySendDetail(ctx context.Context, message *dalModel.MessageSms) (err error) {
	date := time.UnixMilli(message.CreatedAt).Format("20060102")
	r, err := ali.cli.QuerySendDetail(&sms.QuerySendDetailReq{
		Phone: message.Phone,
		BizId: *message.BizID,
		Data:  date,
	})
	if err != nil {
		return err
	}
	if r.Common.Code != "OK" {
		return errors.New(r.Common.Message)
	}
	if r.Status == sms.SendStatusWait {
		return
	}
	message.ErrCode = r.ErrCode
	message.ReceiveDate = r.ReceiveDate
	message.SendDate = r.SendDate
	message.Content = r.Content
	message.MsgStatus = ali.convertStatus(r.Status)
	if err := ali.store.Update(ctx, message); err != nil {
		klog.Errorf("save message to store error: %v", err)
		return err
	}
	return nil
}

func (ali *AliSmsSender) GetProvider() base.SmsProvider {
	return base.SmsProvider_SMS_PROVIDER_ALIYUN
}

func (ali *AliSmsSender) GetConfig() *config.SmsProvider {
	return ali.conf
}

func (ali *AliSmsSender) Allow(api SenderApi) bool {
	return ali.limiters[api].Allow()
}

func (ali *AliSmsSender) Wait(ctx context.Context, api SenderApi) error {
	return ali.limiters[api].Wait(ctx)
}

func (ali *AliSmsSender) convertStatus(s sms.SendStatus) int16 {
	switch s {
	case sms.SendStatusWait:
		return int16(base.MessageStatus_MESSAGE_STATUS_SENDING)
	case sms.SendStatusFailed:
		return int16(base.MessageStatus_MESSAGE_STATUS_FAILED)
	case sms.SendStatusSuccess:
		return int16(base.MessageStatus_MESSAGE_STATUS_SUCCESS)
	default:
		return int16(base.MessageStatus_MESSAGE_STATUS_UNSPECIFIED)
	}
}
