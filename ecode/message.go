package ecode

import "github.com/byteflowing/go-common/ecode"

// 100 - 999 message模块

var (
	ErrCaptchaNotExist      = ecode.NewCode(100, "ERR_CAPTCHA_NOT_EXIST")       // 验证码不存在
	ErrPhoneNotMatch        = ecode.NewCode(101, "ERR_PHONE_NOT_MATCH")         // 手机不匹配
	ErrEmailNotMatch        = ecode.NewCode(102, "ERR_EMAIL_NOT_MATCH")         // 邮箱不匹配
	ErrCaptchaTooManyErrors = ecode.NewCode(103, "ERR_CAPTCHA_TOO_MANY_ERRORS") // 验证码失败次数过多
	ErrCaptchaMisMatch      = ecode.NewCode(104, "ERR_CAPTCHA_MISMATCH")        // 验证码不匹配
	ErrPhoneIsEmpty         = ecode.NewCode(105, "ERR_PHONE_IS_EMPTY")          // 手机号码为空
)
