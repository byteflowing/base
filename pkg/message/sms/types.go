package sms

type SendMessageReq struct {
	PhoneNumber   string
	SignName      string
	TemplateCode  string
	TemplateParam map[string]string
}

type SendCaptchaReq struct {
	PhoneNumber   string
	SignName      string
	TemplateCode  string
	TemplateParam map[string]string
	Captcha       string
	Vendor        Vendor
}

type VerifyCaptchaReq struct {
	Token   string
	Captcha string
}
