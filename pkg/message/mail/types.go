package mail

import "github.com/byteflowing/go-common/mail"

type Vendor string

const (
	VendorAli Vendor = "ali"
)

type SendCaptchaReq struct {
	From        *mail.Address    // 发件人信息
	To          *mail.Address    // 收件人信息
	Subject     string           // 邮件标题
	Captcha     string           // 验证码内容
	ContentType mail.ContentType // 邮件类型，文本/html
	Content     string           // 邮件内容，调用方渲染，里面要包含Captcha
	Vendor      Vendor           // 邮件供应商
}

type VerifyCaptchaReq struct {
	Token   string
	Captcha string
}
