package sms

import (
	"context"
	"errors"
	"fmt"

	limiterv1 "github.com/byteflowing/proto/gen/go/limiter/v1"
	typesv1 "github.com/byteflowing/proto/gen/go/types/v1"
	"google.golang.org/protobuf/proto"

	"github.com/byteflowing/base/app/message/internal/queue"
	"github.com/byteflowing/base/pkg/limiter"
	"github.com/byteflowing/base/pkg/validator"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
)

const (
	accountQuotaName     = "sms_account_quota"
	phoneQuotaName       = "sms_phone_quota"
	accountRuleKeyFormat = "%s:%d:%s"
)

type ISms interface {
	SendSms(ctx context.Context, req *msgv1.SendSmsReq) (resp *msgv1.SendSmsResp, err error)
	SendSmsWithoutLimit(ctx context.Context, req *msgv1.SendSmsWithoutLimitReq) (resp *msgv1.SendSmsResp, err error)
}

type IProvider interface {
	SendSingleSms(ctx context.Context, req *msgv1.SendSmsReq) (err error)
}

type Sms struct {
	cfg        *msgv1.SmsConfig
	providers  map[enumsv1.SmsVendor]map[string]IProvider
	quota      *limiter.Quota
	validators *validator.Chain[*msgv1.SendSmsReq, *msgv1.SendSmsResp]
	queue      queue.Queue
	cancel     context.CancelFunc
}

func New(
	cfg *configv1.Config,
	quota *limiter.Quota,
	queue queue.Queue,
) ISms {
	if cfg == nil || cfg.MessageConfig == nil {
		panic("config is nil")
	}
	if cfg.MessageConfig.Mail == nil {
		return UnimplementedSmsService{}
	}
	smsConfig := cfg.MessageConfig.Sms
	provider := make(map[enumsv1.SmsVendor]map[string]IProvider, len(smsConfig.Providers))
	for _, v := range smsConfig.Providers {
		_, ok := provider[v.GetVendor()]
		if !ok {
			provider[v.GetVendor()] = make(map[string]IProvider)
		}
		provider[v.GetVendor()][v.Account] = newProvider(v)
	}
	sms := &Sms{
		cfg:       smsConfig,
		providers: provider,
		quota:     quota,
		queue:     queue,
	}
	sms.validators = validator.NewChain[*msgv1.SendSmsReq, *msgv1.SendSmsResp](sms.isAllowed)
	sms.validators.Add(accountQuotaName, validator.ValidateFunc[*msgv1.SendSmsReq, *msgv1.SendSmsResp](sms.checkAccountQuota))
	sms.validators.Add(phoneQuotaName, validator.ValidateFunc[*msgv1.SendSmsReq, *msgv1.SendSmsResp](sms.checkPhoneQuota))
	sms.queue.Register(context.Background(), sms.cfg.Topic, sms.consumeHandler)
	return sms
}

func (s *Sms) SendSms(ctx context.Context, req *msgv1.SendSmsReq) (resp *msgv1.SendSmsResp, err error) {
	resp, err = s.validators.Check(ctx, req)
	if err != nil || resp != nil {
		return resp, err
	}
	err = s.queue.Publish(ctx, s.cfg.Topic, req)
	return
}

// SendSmsWithoutLimit 发送消息，不限流手机号，如果当前QPS超过了供应商配置将阻塞等待
// ！！！注意：需要将quota的配置问中的LimitType设置为“LIMITER_TYPE_RATE”方可支持
func (s *Sms) SendSmsWithoutLimit(ctx context.Context, req *msgv1.SendSmsWithoutLimitReq) (resp *msgv1.SendSmsResp, err error) {
	if req.Wait {
		quota, err := s.quota.GetIQuota(ctx, s.getAccountRuleKey(req.Req))
		if err != nil {
			return nil, err
		}
		rateLimiter, ok := quota.(*limiter.RateLimiterQuota)
		if !ok {
			return nil, errors.New("quota is not a rate limiter")
		}
		if err = rateLimiter.Wait(ctx); err != nil {
			return nil, err
		}
		err = s.queue.Publish(ctx, s.cfg.Topic, req)
	} else {
		resp, err = s.validators.CheckWithNames(ctx, req.Req, accountQuotaName)
		if err != nil || resp != nil {
			return resp, err
		}
		err = s.queue.Publish(ctx, s.cfg.Topic, req)
	}
	return
}

func (s *Sms) consumeHandler(ctx context.Context, msg []byte) error {
	m := &msgv1.SendSmsReq{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return err
	}
	p, err := s.getProvider(m)
	if err != nil {
		return err
	}
	return p.SendSingleSms(ctx, m)
}

func (s *Sms) getProvider(req *msgv1.SendSmsReq) (provider IProvider, err error) {
	ps, ok := s.providers[req.GetVendor()]
	if !ok {
		return nil, errors.New("provider not exist")
	}
	p, ok := ps[req.GetAccount()]
	if !ok {
		return nil, errors.New("account not exist")
	}
	return p, nil
}

func (s *Sms) getAccountRuleKey(req *msgv1.SendSmsReq) string {
	// ${QUOTA_PREFIX}:${VENDOR}:${ACCOUNT}
	return fmt.Sprintf(accountRuleKeyFormat, s.cfg.QuotaPrefix, req.Vendor, req.Account)
}

func (s *Sms) getPhoneNumberRuleKey(req *msgv1.SendSmsReq) string {
	return s.cfg.QuotaPrefix
}

func (s *Sms) getFullPhoneNumber(phone *typesv1.PhoneNumber) string {
	return phone.GetCountryCode() + phone.GetNumber()
}

func (s *Sms) isAllowed(result *msgv1.SendSmsResp) bool {
	if result == nil {
		return true
	}
	return result.Rule.Allowed
}

func (s *Sms) checkQuota(ctx context.Context, ruleKey, target string) (*msgv1.SendSmsResp, error) {
	rule, err := s.quota.Take(ctx, &limiterv1.RuleParam{
		Target:  target,
		RuleKey: ruleKey,
	})
	if err != nil {
		return nil, err
	}
	if rule != nil && !rule.Allowed {
		resp := &msgv1.SendSmsResp{
			Rule: rule,
		}
		return resp, nil
	}
	return nil, nil
}

func (s *Sms) checkAccountQuota(ctx context.Context, req *msgv1.SendSmsReq) (*msgv1.SendSmsResp, error) {
	return s.checkQuota(ctx, s.getAccountRuleKey(req), req.Account)
}

func (s *Sms) checkPhoneQuota(ctx context.Context, req *msgv1.SendSmsReq) (*msgv1.SendSmsResp, error) {
	return s.checkQuota(ctx, s.getPhoneNumberRuleKey(req), s.getFullPhoneNumber(req.PhoneNumber))
}

func newProvider(c *msgv1.SmsProvider) IProvider {
	vendor := c.GetVendor()
	switch vendor {
	case enumsv1.SmsVendor_SMS_VENDOR_ALIYUN:
		return NewAli(c)
	case enumsv1.SmsVendor_SMS_VENDOR_VOLC:
		return NewVolc(c)
	}
	panic("unsupported vendor type: " + c.Vendor.String())
}
