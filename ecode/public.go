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
)
