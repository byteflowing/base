package captcha

import (
	"context"
	"fmt"
	"strings"

	"github.com/byteflowing/base/biz/config"
	"github.com/byteflowing/base/biz/constant"
	"github.com/byteflowing/base/biz/pkg/message/limiter"
	"github.com/byteflowing/base/biz/pkg/message/model"
	"github.com/byteflowing/base/biz/pkg/message/sms"
	"github.com/byteflowing/base/kitex_gen/base"
	"github.com/byteflowing/go-common/idx"
	"github.com/byteflowing/go-common/redis"
)

type SmsCaptcha struct {
	conf               *config.SmsConfig
	smsManager         *sms.Manager
	slidingLimiters    map[base.CaptchaType]limiter.Limiter
	prefixes           map[base.CaptchaType]string
	dailyLimiter       limiter.Limiter
	errTryLimiter      limiter.Limiter
	dailyLimiterPrefix string
	store              Store
}

type SmsCaptchaOption struct {
	SmsConfig    *config.SmsConfig
	SmsManager   *sms.Manager
	CaptchaStore Store
	RDB          *redis.Redis
}

func NewSmsCaptcha(opts *SmsCaptchaOption) Captcha {
	conf := opts.SmsConfig.Captcha
	slidingLimiters := make(map[base.CaptchaType]limiter.Limiter, len(conf.Limits))
	prefixes := make(map[base.CaptchaType]string, len(conf.Limits))
	for _, v := range conf.Limits {
		keyPrefix := fmt.Sprintf("%s:%d:%d", conf.KeyPrefix, int(base.MessageSender_MESSAGE_SENDER_SMS), int(v.ToCaptchaType()))
		slidingLimiters[v.ToCaptchaType()] = redis.NewLimiter(opts.RDB, keyPrefix, v.ToWindows())
		prefixes[v.ToCaptchaType()] = keyPrefix
	}
	dailyLimiterPrefix := fmt.Sprintf("%s:%s", conf.KeyPrefix, "daily")
	errTryPrefix := fmt.Sprintf("%s:%s", conf.KeyPrefix, "err")
	return &SmsCaptcha{
		smsManager:      opts.SmsManager,
		conf:            opts.SmsConfig,
		store:           opts.CaptchaStore,
		slidingLimiters: slidingLimiters,
		prefixes:        prefixes,
		dailyLimiter:    limiter.NewDailyLimiter(opts.RDB, dailyLimiterPrefix, opts.SmsConfig.Captcha.DailyLimit),
		errTryLimiter:   redis.NewLimiter(opts.RDB, errTryPrefix, []*redis.Window{conf.ErrTryLimit.ToWindow()}),
	}
}

func (s *SmsCaptcha) Send(ctx context.Context, req *base.SendCaptchaReq) (resp *base.SendCaptchaResp, err error) {
	slidingLimiter, ok := s.slidingLimiters[req.CaptchaType]
	if !ok {
		return nil, constant.ErrNotImplemented
	}
	ok, window, err := slidingLimiter.Allow(ctx, *req.Phone)
	if err != nil {
		return nil, err
	}
	if !ok {
		resp = &base.SendCaptchaResp{Tag: &window.Tag}
		return resp, constant.ErrResourceLimited
	}
	ok, _, err = s.dailyLimiter.Allow(ctx, *req.Phone)
	if err != nil {
		return nil, err
	}
	if !ok {
		resp = &base.SendCaptchaResp{Tag: &s.conf.Captcha.DailyLimitTag}
		return resp, constant.ErrResourceLimited
	}
	sender, err := s.smsManager.GetSender(*req.SmsProvider)
	if err != nil {
		return nil, err
	}
	if ok := sender.Allow(sms.ApiSendMessage); !ok {
		return nil, constant.ErrSystemBusy
	}
	// TODO: 从ctx解析用户id
	if err = sender.SendMessage(ctx, &model.SendSmsMessageReq{
		SmsMessage: &model.SmsMessage{
			Type:           base.MessageType_MESSAGE_TYPE_CAPTCHA,
			Provider:       *req.SmsProvider,
			CaptchaType:    &req.CaptchaType,
			SenderID:       0,
			PhoneNumber:    *req.Phone,
			SignName:       *req.Sign,
			TemplateCode:   *req.Template,
			TemplateParams: req.Params,
		},
	}); err != nil {
		return nil, err
	}
	prefix := s.prefixes[req.CaptchaType]
	token := idx.UUIDv4()
	key := fmt.Sprintf("%s:%s", prefix, token)
	err = s.store.Save(ctx, key, req.CaptchaContent, s.conf.Captcha.KeepingToDuration())
	return
}

func (s *SmsCaptcha) Verify(ctx context.Context, req *base.VerifyCaptchaReq) (bool, error) {
	prefix, ok := s.prefixes[req.CaptchaType]
	if !ok {
		return false, constant.ErrNotImplemented
	}
	key := fmt.Sprintf("%s:%s", prefix, req.Token)
	v, ok, err := s.store.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, constant.ErrNoCaptcha
	}
	captcha := req.Captcha
	if !s.conf.Captcha.CaseSensitive {
		captcha = strings.ToLower(captcha)
	}
	if v != captcha {
		ok, _, err := s.errTryLimiter.Allow(ctx, req.Token)
		if err != nil {
			return false, err
		}
		if !ok {
			if err := s.store.Delete(ctx, key); err != nil {
				return false, err
			}
			return false, constant.ErrCaptchaTryTooMany
		}
		return false, nil
	}
	if err = s.store.Delete(ctx, key); err != nil {
		return false, err
	}
	return true, nil
}
