package sms

import "github.com/byteflowing/base/pkg/message/captcha"

type Vendor string

const (
	VendorAli  Vendor = "ali"
	VendorVolc Vendor = "volc"
)

type Config struct {
	Providers []*ProviderConfig
	Captcha   *captcha.Config
}

type ProviderConfig struct {
	Account        string  // 短信账号，volc需要
	Vendor         string  // 短信供应商 "ali", "volc"
	AccessKey      string  // ak
	SecretKey      string  // sk
	SecurityToken  *string // sts场景的security Token
	SendMessageQPS uint32  // 发送接口qps
}

func (p *ProviderConfig) GetVendor() Vendor {
	switch p.Vendor {
	case "ali":
		return VendorAli
	case "volc":
		return VendorVolc
	}
	panic("unsupported vendor type:" + p.Vendor)
}
