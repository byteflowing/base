package constant

import "github.com/byteflowing/go-common/ecode"

// 为了适配前端的i18n这里的errMsg全部使用大写英文且不能重复
// 0 - 99 公共错误
// 100 - 999 message模块
// 1000 - 1999 user模块
var (
	StatusOK                   = ecode.NewCode(0, "OK")                              // 正常
	ErrParams                  = ecode.NewCode(1, "ERR_PARAMS")                      // 参数错误
	ErrInternal                = ecode.NewCode(2, "ERR_INTERNAL")                    // 内部错误
	ErrNotFound                = ecode.NewCode(3, "ERR_NOT_FOUND")                   // 资源不存在
	ErrUnauthorized            = ecode.NewCode(4, "ERR_UNAUTHORIZED")                // 未认证
	ErrPermission              = ecode.NewCode(5, "ERR_PERMISSION")                  // 无权限
	ErrTimeout                 = ecode.NewCode(6, "ERR_TIMEOUT")                     // 超时
	ErrSystemBusy              = ecode.NewCode(7, "ERR_SYSTEM_BUSY")                 // 系统繁忙
	ErrResourceLimited         = ecode.NewCode(8, "ERR_RESOURCE_LIMITED")            // 资源被限制使用
	ErrNotImplemented          = ecode.NewCode(9, "ERR_NOT_IMPLEMENTED")             // 功能未实现
	ErrTooManyCaptcha          = ecode.NewCode(100, "ERR_TOO_MANY_CAPTCHA")          // 验证码发送太频繁
	ErrCaptchaReach10MinLimits = ecode.NewCode(101, "ERR_CAPTCHA_REACH_10MIN_LIMIT") // 验证码超过十分钟最大条数限制
	ErrCaptchaReach1DayLimits  = ecode.NewCode(102, "ERR_CAPTCHA_REACH_1DAY_LIMIT")  // 验证码超过一天最大条数限制
	ErrCaptchaNotExists        = ecode.NewCode(103, "ERR_CAPTCHA_NOT_EXISTS")        // 验证码不存在或者已过期
	ErrCaptchaVerifyTooMany    = ecode.NewCode(104, "ERR_CAPTCHA_VERIFY_TOO_MANY")   // 验证码尝试次数过多
	ErrCaptchaMisMatch         = ecode.NewCode(105, "ERR_CAPTCHA_MISMATCH")          // 验证码不正确
	ErrNoCaptcha               = ecode.NewCode(106, "ERR_NO_CAPTCHA")                // 无验证
	ErrCaptchaTryTooMany       = ecode.NewCode(107, "ERR_CAPTCHA_TRY_TOO_MANY")      // 验证码尝试次数过多
)
