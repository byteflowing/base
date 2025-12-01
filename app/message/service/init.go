package service

import (
	"fmt"
	"slices"

	"github.com/byteflowing/base/app/message/provider/mail"
	"github.com/byteflowing/base/app/message/provider/sms"
	"github.com/byteflowing/base/pkg/captcha"
	"github.com/byteflowing/base/pkg/quota"
	"github.com/byteflowing/base/pkg/redis"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	msgv1 "github.com/byteflowing/proto/gen/go/msg/v1"
)

var (
	smsInterface = []enumv1.MessageInterface{
		enumv1.MessageInterface_MESSAGE_INTERFACE_SMS_SEND_SINGLE_MESSAGE,
		enumv1.MessageInterface_MESSAGE_INTERFACE_SMS_QUERY_SEND_DETAILS,
		enumv1.MessageInterface_MESSAGE_INTERFACE_SMS_QUERY_SEND_STATISTICS,
	}
	emailInterface = []enumv1.MessageInterface{
		enumv1.MessageInterface_MESSAGE_INTERFACE_MAIL_SEND_SINGLE_MESSAGE,
		enumv1.MessageInterface_MESSAGE_INTERFACE_MAIL_SEND_STATISTICS,
		enumv1.MessageInterface_MESSAGE_INTERFACE_MAIL_SEND_DETAILS,
		enumv1.MessageInterface_MESSAGE_INTERFACE_MAIL_TRACK_LIST,
	}
)

func initSms(cfg *configv1.Config, rdb *redis.Redis, s *MessageService) {
	if cfg.Message.Sms == nil {
		return
	}
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS
	s.smsProviders = sms.NewSms(cfg.Message.Sms)
	s.captcha[t] = newCaptcha(rdb, cfg.Message.Captcha.SmsCaptcha, cfg.Message.Captcha.Prefix, t)
	s.slidingRules[t] = convertSlidingWindows(cfg.Message.Captcha.SmsCaptcha.Quota)
	s.rateLimiterCapacity[t] = getSmsRateLimiterCapacities(cfg.Message.Sms.Providers)
	s.smsAccountsMapping = getSmsAccountMappings(cfg.Message.Sms.Providers)
	s.queue.RegisterHandler(s.getTaskName(taskTypeSend, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS), s.sendSms)
}

func initMail(cfg *configv1.Config, rdb *redis.Redis, s *MessageService) {
	if cfg.Message.Mail == nil {
		return
	}
	t := enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL
	s.mailProviders = mail.NewMail(cfg.Message.Mail)
	s.captcha[t] = newCaptcha(rdb, cfg.Message.Captcha.MailCaptcha, cfg.Message.Captcha.Prefix, t)
	s.slidingRules[t] = convertSlidingWindows(cfg.Message.Captcha.MailCaptcha.Quota)
	s.rateLimiterCapacity[t] = getMailRateLimiterCapacities(cfg.Message.Mail.Providers)
	s.mailAccountsMapping = getMailAccountMappings(cfg.Message.Mail.Providers)
	s.queue.RegisterHandler(s.getTaskName(taskTypeSend, enumv1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL), s.sendMail)
}

func newCaptcha(rdb *redis.Redis, cfg *msgv1.CaptchaItem, prefix string, sender enumv1.MessageSenderType) *captcha.MessageCaptcha {
	return captcha.NewMessageCaptcha(rdb, &captcha.Config{
		Prefix:              fmt.Sprintf(captchaPrefixFormat, prefix, sender),
		MaxTries:            int(cfg.MaxTries),
		Length:              int(cfg.Length),
		CaptchaTTL:          cfg.Ttl.AsDuration(),
		CaptchaCombinations: cfg.Masks,
	})
}

func convertSlidingWindows(rules []*msgv1.CaptchaQuota) map[enumv1.MessageSceneType][]*quota.SlidingRule {
	slidingRules := make(map[enumv1.MessageSceneType][]*quota.SlidingRule)
	for _, rule := range rules {
		slidingRules[rule.Scene] = append(slidingRules[rule.Scene], &quota.SlidingRule{
			Quota:  int(rule.Quota),
			Window: rule.Window.AsDuration(),
		})
	}
	return slidingRules
}

func getSmsRateLimiterCapacities(providers []*msgv1.SmsProvider) map[enumv1.MessageSenderVendor]map[string]map[enumv1.MessageInterface]*rateLimiterConfig {
	capacities := make(map[enumv1.MessageSenderVendor]map[string]map[enumv1.MessageInterface]*rateLimiterConfig)
	for _, provider := range providers {
		_, ok := capacities[provider.Vendor]
		if !ok {
			capacities[provider.Vendor] = make(map[string]map[enumv1.MessageInterface]*rateLimiterConfig)
		}
		capacities[provider.Vendor][provider.Account] = make(map[enumv1.MessageInterface]*rateLimiterConfig)
		if len(provider.Quota) != len(smsInterface) {
			panic("sms quota config is not correct")
		}
		for _, i := range provider.Quota {
			if !slices.Contains(smsInterface, i.Interface) {
				panic("sms interface not supported")
			}
			capacities[provider.Vendor][provider.Account][i.Interface] = &rateLimiterConfig{
				quota:    int64(i.Quota),
				interval: i.Interval.AsDuration(),
			}
		}
	}
	return capacities
}

func getMailRateLimiterCapacities(providers []*msgv1.MailProvider) map[enumv1.MessageSenderVendor]map[string]map[enumv1.MessageInterface]*rateLimiterConfig {
	capacities := make(map[enumv1.MessageSenderVendor]map[string]map[enumv1.MessageInterface]*rateLimiterConfig)
	for _, provider := range providers {
		_, ok := capacities[provider.Vendor]
		if !ok {
			capacities[provider.Vendor] = make(map[string]map[enumv1.MessageInterface]*rateLimiterConfig)
		}
		capacities[provider.Vendor][provider.Account] = make(map[enumv1.MessageInterface]*rateLimiterConfig)
		if len(provider.Quota) != len(emailInterface) {
			panic("quota config is not correct")
		}
		for _, i := range provider.Quota {
			if !slices.Contains(emailInterface, i.Interface) {
				panic("email interface not supported")
			}
			capacities[provider.Vendor][provider.Account][i.Interface] = &rateLimiterConfig{
				quota:    int64(i.Quota),
				interval: i.Interval.AsDuration(),
			}
		}
	}
	return capacities
}

func getSmsAccountMappings(providers []*msgv1.SmsProvider) []*msgv1.VendorAccountMapping {
	var mappings []*msgv1.VendorAccountMapping
	vendorAccountMappings := make(map[enumv1.MessageSenderVendor][]string)
	for _, provider := range providers {
		vendorAccountMappings[provider.Vendor] = append(vendorAccountMappings[provider.Vendor], provider.Account)
	}
	for vendor, accounts := range vendorAccountMappings {
		mappings = append(mappings, &msgv1.VendorAccountMapping{
			Vendor:  vendor,
			Account: accounts,
		})
	}
	return mappings
}

func getMailAccountMappings(providers []*msgv1.MailProvider) []*msgv1.VendorAccountMapping {
	var mappings []*msgv1.VendorAccountMapping
	vendorAccountMappings := make(map[enumv1.MessageSenderVendor][]string)
	for _, provider := range providers {
		vendorAccountMappings[provider.Vendor] = append(vendorAccountMappings[provider.Vendor], provider.Account)
	}
	for vendor, accounts := range vendorAccountMappings {
		mappings = append(mappings, &msgv1.VendorAccountMapping{
			Vendor:  vendor,
			Account: accounts,
		})
	}
	return mappings
}
