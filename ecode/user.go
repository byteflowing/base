package ecode

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUserAuthInvalid  = status.Error(codes.PermissionDenied, "ERR_USER_AUTH_INVALID")  // 当前认证不可用
	ErrUserTokenInvalid = status.Error(codes.PermissionDenied, "ERR_USER_TOKEN_INVALID") // 登录信息不可用
	ErrUserDisabled     = status.Error(codes.PermissionDenied, "ERR_USER_DISABLED")      // 用户被禁用
	ErrUserNotFound     = status.Error(codes.NotFound, "ERR_USER_NOT_FOUND")             // 用户不存在
)
