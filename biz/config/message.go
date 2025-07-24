package config

import (
	"time"

	"github.com/byteflowing/base/kitex_gen/base"
	"github.com/byteflowing/go-common/redis"
)

type MessageConfig struct {
	Active bool       // 是否启用该模块
	Sms    *SmsConfig // 短信配置
}

type SmsConfig struct {
	SaveMessage         bool           // 是否保存发送的消息
	QueryDetailInterval uint32         // 向短信提供商回查短信详情时间间隔单位s
	Providers           []*SmsProvider // 短信供应商
	Captcha             *CaptchaConfig // 验证码配置
}

type SmsProvider struct {
	Provider             string  // 短信供应商 "aliyun"
	AccessKey            string  // ak
	SecretKey            string  // sk
	SecurityKey          *string // sts场景的security token
	SendMessageLimit     uint32  // 发送短信接口的供应商限流
	QuerySendDetailLimit uint32  // 查询短信详情的供应商限流
}

func (c *SmsProvider) GetProvider() base.SmsProvider {
	switch c.Provider {
	case "aliyun":
		return base.SmsProvider_SMS_PROVIDER_ALIYUN
	default:
		return base.SmsProvider_SMS_PROVIDER_UNSPECIFIED
	}
}

type CaptchaConfig struct {
	KeyPrefix     string         // 验证码在redis中存储的前缀
	Keeping       uint32         // 验证码保留时长，单位s
	CaseSensitive bool           // 大小写敏感
	DailyLimit    uint32         // 一个手机号码一自然日的发送限制条数(00:00 - 23:59:59)
	DailyLimitTag string         // 全局限流tag
	ErrTryLimit   *LimitRule     // 可以尝试几次验证，超过后删除验证码
	Limits        []*LimitConfig // 验证码短信限流规则
}

func (c *CaptchaConfig) KeepingToDuration() time.Duration {
	return time.Duration(c.Keeping) * time.Second
}

type LimitConfig struct {
	CaptchaType string       // 验证码类型，"login", "other"
	Rules       []*LimitRule // 限流规则
}

func (c *LimitConfig) ToWindows() []*redis.Window {
	var windows []*redis.Window
	for _, rule := range c.Rules {
		windows = append(windows, rule.ToWindow())
	}
	return windows
}

func (c *LimitConfig) ToCaptchaType() base.CaptchaType {
	switch c.CaptchaType {
	case "login":
		return base.CaptchaType_CAPTCHA_TYPE_LOGIN
	case "other":
		return base.CaptchaType_CAPTCHA_TYPE_OTHER
	default:
		return base.CaptchaType_CAPTCHA_TYPE_UNSPECIFIED
	}
}

type LimitRule struct {
	Duration uint32 // 短信发送限制时间周期，单位s
	Limit    uint32 // 周期内的限制次数
	Tag      string // 标记，用户业务识别方便错误提示
}

func (l *LimitRule) ToWindow() *redis.Window {
	return &redis.Window{
		Duration: l.GetDuration(),
		Limit:    l.Limit,
		Tag:      l.Tag,
	}
}

func (l *LimitRule) GetDuration() time.Duration {
	return time.Duration(l.Duration) * time.Second
}
