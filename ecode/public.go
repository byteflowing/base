package ecode

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInternal         = status.New(codes.Internal, "ERR_INTERNAL").Err()                   // 内部错误
	ErrNotFound         = status.New(codes.NotFound, "ERR_NOT_FOUND").Err()                  // 资源不存在
	ErrUnauthenticated  = status.New(codes.Unauthenticated, "ERR_UNAUTHORIZED").Err()        // 未认证
	ErrPermission       = status.New(codes.PermissionDenied, "ERR_PERMISSION").Err()         // 无权限
	ErrTimeout          = status.New(codes.DeadlineExceeded, "ERR_TIMEOUT").Err()            // 超时
	ErrTooManyRequests  = status.New(codes.ResourceExhausted, "ERR_TOO_MANY_REQUESTS").Err() // 请求过多
	ErrUnImplemented    = status.New(codes.Unimplemented, "ERR_UNIMPLEMENTED").Err()         // 未实现
	ErrParams           = status.New(codes.InvalidArgument, "ERR_PARAMS").Err()              // 参数错误
	ErrPhoneNotMatch    = status.New(codes.InvalidArgument, "ERR_PHONE_NOT_MATCH").Err()     // 手机不匹配
	ErrEmailNotMatch    = status.New(codes.InvalidArgument, "ERR_EMAIL_NOT_MATCH").Err()     // 邮箱不匹配
	ErrPhoneIsEmpty     = status.New(codes.InvalidArgument, "ERR_PHONE_IS_EMPTY").Err()      // 手机号码为空
	ErrEmailIsEmpty     = status.New(codes.InvalidArgument, "ERR_EMAIL_IS_EMPTY").Err()      // 邮箱为空
	ErrPhoneNotExist    = status.New(codes.InvalidArgument, "ERR_PHONE_NOT_EXIST").Err()     // 手机号码不存在
	ErrEmailNotExist    = status.New(codes.InvalidArgument, "ERR_EMAIL_NOT_EXIST").Err()     // 邮箱不存在
	ErrPhoneExists      = status.New(codes.InvalidArgument, "ERR_PHONE_EXISTS").Err()        // 手机已存在
	ErrEmailExists      = status.New(codes.InvalidArgument, "ERR_EMAIL_EXISTS").Err()        // 邮箱已存在
	ErrEmailAlreadyBind = status.New(codes.AlreadyExists, "ERR_EMAIL_ALREADY_BIND").Err()    // 邮箱已绑定
	ErrPhoneAlreadyBind = status.New(codes.AlreadyExists, "ERR_PHONE_ALREADY_BIND").Err()    // 手机已绑定
)
