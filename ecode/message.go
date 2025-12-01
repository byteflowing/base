package ecode

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrMsgReachDailyQuota    = status.Error(codes.ResourceExhausted, "ERR_REACH_DAILY_QUOTA") // 请求量达到用户日上限
	ErrMsgSenderUnsupported  = status.Error(codes.Internal, "ERR_SENDER_UNSUPPORTED")         // 当前发送程序不支持
	ErrMsgVendorUnsupported  = status.Error(codes.Internal, "ERR_VENDOR_UNSUPPORTED")         // 当前供应商不支持
	ErrMsgAccountUnsupported = status.Error(codes.Internal, "ERR_ACCOUNT_UNSUPPORTED")        // 当前账号不支持
	ErrMsgReachQuota         = status.Error(codes.ResourceExhausted, "ERR_REACH_QUOTA")       // 触发限流
	ErrMsgSceneUnsupported   = status.Error(codes.Internal, "ERR_SCENE_UNSUPPORTED")          // 当前场景不支持
)
