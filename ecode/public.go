package ecode

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInternal              = status.Error(codes.Internal, "ERR_INTERNAL")                        // 内部错误
	ErrNotFound              = status.Error(codes.NotFound, "ERR_NOT_FOUND")                       // 资源不存在
	ErrPhoneNotExist         = status.Error(codes.NotFound, "ERR_PHONE_NOT_EXIST")                 // 手机号码不存在
	ErrEmailNotExist         = status.Error(codes.NotFound, "ERR_EMAIL_NOT_EXIST")                 // 邮箱不存在
	ErrUnauthenticated       = status.Error(codes.Unauthenticated, "ERR_UNAUTHORIZED")             // 未认证
	ErrPermission            = status.Error(codes.PermissionDenied, "ERR_PERMISSION")              // 无权限
	ErrTimeout               = status.Error(codes.DeadlineExceeded, "ERR_TIMEOUT")                 // 超时
	ErrNoResource            = status.Error(codes.ResourceExhausted, "ERR_NO_RESOURCE")            // 系统资源耗尽
	ErrTooManyRequests       = status.Error(codes.ResourceExhausted, "ERR_TOO_MANY_REQUESTS")      // 请求过多
	ErrMapReachIntervalLimit = status.Error(codes.ResourceExhausted, "ERR_MAP_REACH_SECOND_LIMIT") // 请求量已达到单位时间上限
	ErrMapReachDailyLimit    = status.Error(codes.ResourceExhausted, "ERR_MAP_REACH_DAILY_LIMIT")  // 每日调用量已达到上限
	ErrLockFailed            = status.Error(codes.FailedPrecondition, "ERR_LOCK_FAILED")           // 上锁失败
	ErrUnLockFailed          = status.Error(codes.FailedPrecondition, "ERR_UNLOCK_FAILED")         // 解锁失败
	ErrUnImplemented         = status.Error(codes.Unimplemented, "ERR_UNIMPLEMENTED")              // 未实现
	ErrParams                = status.Error(codes.InvalidArgument, "ERR_PARAMS")                   // 参数错误
	ErrPhoneNotMatch         = status.Error(codes.InvalidArgument, "ERR_PHONE_NOT_MATCH")          // 手机不匹配
	ErrEmailNotMatch         = status.Error(codes.InvalidArgument, "ERR_EMAIL_NOT_MATCH")          // 邮箱不匹配
	ErrPhoneIsEmpty          = status.Error(codes.InvalidArgument, "ERR_PHONE_IS_EMPTY")           // 手机号码为空
	ErrEmailIsEmpty          = status.Error(codes.InvalidArgument, "ERR_EMAIL_IS_EMPTY")           // 邮箱为空
	ErrCaptchaTokenIsEmpty   = status.Error(codes.InvalidArgument, "ERR_CAPTCHA_TOKEN_IS_EMPTY")   // 验证码token为空
	ErrJwtSignMethodMismatch = status.Error(codes.InvalidArgument, "ERR_JWT_SIGN_METHOD_MISMATCH") // token签名算法不匹配
	ErrJwtIssuerMismatch     = status.Error(codes.InvalidArgument, "ERR_JWT_ISSUER_MISMATCH")      // token签发人不匹配
	ErrJwtTokenTypeMismatch  = status.Error(codes.InvalidArgument, "ERR_JWT_TOKEN_TYPE_MISMATCH")  // token类型不匹配
	ErrJwtInvalidToken       = status.Error(codes.InvalidArgument, "ERR_JWT_INVALID_TOKEN")        // token不可用
	ErrJwtTokenRevoked       = status.Error(codes.PermissionDenied, "ERR_JWT_TOKEN_REVOKED")       // token已禁用
	ErrPhoneCodeExists       = status.Error(codes.AlreadyExists, "ERR_PHONE_CODE_EXISTS")          // 手机码已存在
	ErrPhoneExists           = status.Error(codes.AlreadyExists, "ERR_PHONE_EXISTS")               // 手机已存在
	ErrEmailExists           = status.Error(codes.AlreadyExists, "ERR_EMAIL_EXISTS")               // 邮箱已存在
	ErrEmailAlreadyBind      = status.Error(codes.AlreadyExists, "ERR_EMAIL_ALREADY_BIND")         // 邮箱已绑定
	ErrPhoneAlreadyBind      = status.Error(codes.AlreadyExists, "ERR_PHONE_ALREADY_BIND")         // 手机已绑定
)
