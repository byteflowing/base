package ecode

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCaptchaNotExist      = status.New(codes.InvalidArgument, "ERR_CAPTCHA_NOT_EXIST").Err()         // 没有验证码
	ErrCaptchaTooManyErrors = status.New(codes.ResourceExhausted, "ERR_CAPTCHA_TOO_MANY_ERRORS").Err() // 验证码失败次数过多
	ErrCaptchaMisMatch      = status.New(codes.InvalidArgument, "ERR_CAPTCHA_MISMATCH").Err()          // 验证码不匹配
)
