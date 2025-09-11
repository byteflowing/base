package ecode

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUserNotExist             = status.New(codes.NotFound, "ERR_USER_NOT_EXIST").Err()                        // 用户不存在
	ErrUserNumberNotExist       = status.New(codes.NotFound, "ERR_USER_NUMBER_NOT_EXIST").Err()                 // 用户编号不存在
	ErrUserNumberExists         = status.New(codes.NotFound, "ERR_USER_NUMBER_EXISTS").Err()                    // 用户编号已存在
	ErrUserPasswordNotSet       = status.New(codes.FailedPrecondition, "ERR_USER_PASSWORD_NOT_SET").Err()       // 用户未设置密码
	ErrUserPasswordWrongTooMany = status.New(codes.ResourceExhausted, "ERR_USER_PASSWORD_WRONG_TOO_MANY").Err() // 密码错误次数过多
	ErrUserPasswordMisMatch     = status.New(codes.InvalidArgument, "ERR_USER_PASSWORD_MISMATCH").Err()         // 用户密码不匹配
	ErrUserAuthTypeMisMatch     = status.New(codes.InvalidArgument, "ERR_USER_AUTH_TYPE_MISMATCH").Err()        // 认证类型不匹配
	ErrUserTokenMisMatch        = status.New(codes.InvalidArgument, "ERR_USER_TOKEN_MISMATCH").Err()            // 登录信息不匹配
	ErrUserInvalidToken         = status.New(codes.InvalidArgument, "ERR_USER_INVALID_TOKEN").Err()             // 无效登录信息
	ErrUserCaptchaIsEmpty       = status.New(codes.InvalidArgument, "ERR_USER_CAPTCHA_IS_EMPTY").Err()          // 验证码为空
	ErrUserCaptchaTokenIsEmpty  = status.New(codes.InvalidArgument, "ERR_USER_CAPTCHA_TOKEN_IS_EMPTY").Err()    // 验证码token为空
	ErrUserTokenExpired         = status.New(codes.Unauthenticated, "ERR_USER_TOKEN_EXPIRED").Err()             // 登录信息已失效
	ErrUserTokenRevoked         = status.New(codes.Unauthenticated, "ERR_USER_TOKEN_REVOKED").Err()             // 登录信息已注销
	ErrUserDisabled             = status.New(codes.PermissionDenied, "ERR_USER_DISABLED").Err()                 // 用户被禁用
	ErrUserAuthDisabled         = status.New(codes.PermissionDenied, "ERR_USER_AUTH_DISABLED").Err()            // 登录验证禁用
	ErrUserSignInTooMany        = status.New(codes.ResourceExhausted, "ERR_USER_SIGN_IN_TOO_MANY").Err()        // 登录过于频繁
	ErrUserRefreshTooMany       = status.New(codes.ResourceExhausted, "ERR_USER_REFRESH_TOO_MANY").Err()        // 登录刷新过于频繁
)
