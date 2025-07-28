package ecode

import "github.com/byteflowing/go-common/ecode"

// 100 - 999 message模块

var (
	ErrCaptchaTriesTooMany = ecode.NewCode(100, "ERR_CAPTCHA_TRIES_TOO_MANY")
)
