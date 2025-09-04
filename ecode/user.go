package ecode

import "github.com/byteflowing/go-common/ecode"

// user模块 1000 - 1999

var (
	ErrUserNotExist             = ecode.NewCode(1000, "ERR_USER_NOT_EXIST")               // 用户不存在
	ErrUserDisabled             = ecode.NewCode(1001, "ERR_USER_DISABLED")                // 用户被禁用
	ErrUserPasswordNotSet       = ecode.NewCode(1002, "ERR_USER_PASSWORD_NOT_SET")        // 用户未设置密码
	ErrUserPasswordMisMatch     = ecode.NewCode(1003, "ERR_USER_PASSWORD_MISMATCH")       // 用户密码不匹配
	ErrUserAuthTypeMisMatch     = ecode.NewCode(1004, "ERR_USER_AUTH_TYPE_MISMATCH")      // 认证类型不匹配
	ErrUserTokenMisMatch        = ecode.NewCode(1005, "ERR_USER_TOKEN_MISMATCH")          // 登录信息不匹配
	ErrUserInvalidToken         = ecode.NewCode(1006, "ERR_USER_INVALID_TOKEN")           // 无效登录信息
	ErrUserTokenExpired         = ecode.NewCode(1007, "ERR_USER_TOKEN_EXPIRED")           // 登录信息已失效
	ErrUserTokenRevoked         = ecode.NewCode(1008, "ERR_USER_TOKEN_REVOKED")           // 登录信息已注销
	ErrUserCaptchaIsEmpty       = ecode.NewCode(1009, "ERR_USER_CAPTCHA_IS_EMPTY")        // 验证码为空
	ErrUserCaptchaTokenIsEmpty  = ecode.NewCode(1010, "ERR_USER_CAPTCHA_TOKEN_IS_EMPTY")  // 验证码token为空
	ErrUserAuthDisabled         = ecode.NewCode(1011, "ERR_USER_AUTH_DISABLED")           // 登录验证禁用
	ErrUserSignInTooMany        = ecode.NewCode(1012, "ERR_USER_SIGN_IN_TOO_MANY")        // 登录过于频繁
	ErrUserRefreshTooMany       = ecode.NewCode(1013, "ERR_USER_REFRESH_TOO_MANY")        // 登录刷新过于频繁
	ErrUserNumberNotExist       = ecode.NewCode(1014, "ERR_USER_NUMBER_NOT_EXIST")        // 用户编号不存在
	ErrUserNumberExists         = ecode.NewCode(1015, "ERR_USER_NUMBER_EXISTS")           // 用户编号已存在
	ErrUserPasswordWrongTooMany = ecode.NewCode(1016, "ERR_USER_PASSWORD_WRONG_TOO_MANY") // 密码错误次数过多
)
