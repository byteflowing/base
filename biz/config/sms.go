package config

import (
	"errors"

	"github.com/byteflowing/base/biz/constant"
	aliSms "github.com/byteflowing/go-common/3rd/aliyun/sms"
)

type SmsConfig struct {
	// 短信供应商
	// aliyun
	Provider string
	AliSms   *aliSms.Opts
}

func (c *SmsConfig) GetProvider() constant.SmsProvider {
	switch c.Provider {
	case "aliyun":
		return constant.SmsProviderAliyun
	default:
		return constant.SmsProviderUnknown
	}
}

func (c *SmsConfig) IsValidProvider() bool {
	p := c.GetProvider()
	if p == constant.SmsProviderUnknown {
		return false
	}
	return true
}

func ValidateSmsConfig(c map[string]*SmsConfig) error {
	if c != nil {
		for k, v := range c {
			if !v.IsValidProvider() {
				return errors.New("invalid provider " + k)
			}
			if k != v.Provider {
				return errors.New("provider mismatch")
			}
		}
	}
	return nil
}
