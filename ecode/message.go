package ecode

import "github.com/byteflowing/go-common/ecode"

// 100 - 999 message模块

var (
	ErrCaptchaNotExist      = ecode.NewCode(100, "ERR_CAPTCHA_NOT_EXIST")       // 验证码不存在
	ErrCaptchaTooManyErrors = ecode.NewCode(101, "ERR_CAPTCHA_TOO_MANY_ERRORS") // 验证码失败次数过多
	ErrCaptchaMisMatch      = ecode.NewCode(102, "ERR_CAPTCHA_MISMATCH")        // 验证码不匹配
)
