package mail

import (
	"github.com/byteflowing/base/pkg/message/captcha"
	"github.com/byteflowing/go-common/mail"
)

type Config struct {
	Providers []*Provider
	Captcha   *captcha.Config
}

type Provider struct {
	Vendor         string         // 供应商： "ali"
	LimitDuration  uint64         // 限流的时间间隔单位s
	LimitMax       uint64         // 限流时间间隔内最大请求量
	MaxConnections uint64         // 最大连接数
	SMTP           *mail.SMTPOpts // smtp配置
}

func (p *Provider) GetVendor() Vendor {
	switch p.Vendor {
	case "ali":
		return VendorAli
	}
	return VendorAli
}
