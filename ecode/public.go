package ecode

import "github.com/byteflowing/go-common/ecode"

// 0 - 99 公共错误
var (
	StatusOK           = ecode.NewCode(0, "OK")                    // 正常
	ErrParams          = ecode.NewCode(1, "ERR_PARAMS")            // 参数错误
	ErrInternal        = ecode.NewCode(2, "ERR_INTERNAL")          // 内部错误
	ErrNotFound        = ecode.NewCode(3, "ERR_NOT_FOUND")         // 资源不存在
	ErrUnauthorized    = ecode.NewCode(4, "ERR_UNAUTHORIZED")      // 未认证
	ErrPermission      = ecode.NewCode(5, "ERR_PERMISSION")        // 无权限
	ErrTimeout         = ecode.NewCode(6, "ERR_TIMEOUT")           // 超时
	ErrTooManyRequests = ecode.NewCode(7, "ERR_TOO_MANY_REQUESTS") // 请求过多
	ErrUnImplemented   = ecode.NewCode(8, "ERR_UNIMPLEMENTED")     // 未实现
	ErrPhoneNotMatch   = ecode.NewCode(9, "ERR_PHONE_NOT_MATCH")   // 手机不匹配
	ErrEmailNotMatch   = ecode.NewCode(10, "ERR_EMAIL_NOT_MATCH")  // 邮箱不匹配
	ErrPhoneIsEmpty    = ecode.NewCode(11, "ERR_PHONE_IS_EMPTY")   // 手机号码为空
	ErrEmailIsEmpty    = ecode.NewCode(12, "ERR_EMAIL_IS_EMPTY")   // 邮箱为空
	ErrPhoneNotExist   = ecode.NewCode(13, "ERR_PHONE_NOT_EXIST")  // 手机号码不存在
	ErrEmailNotExist   = ecode.NewCode(14, "ERR_EMAIL_NOT_EXIST")  // 邮箱不存在
	ErrPhoneExists     = ecode.NewCode(15, "ERR_PHONE_EXISTS")     // 手机已存在
	ErrEmailExists     = ecode.NewCode(16, "ERR_EMAIL_EXISTS")     // 邮箱已存在
)
