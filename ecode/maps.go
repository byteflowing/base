package ecode

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrMapsNoResource             = status.Error(codes.ResourceExhausted, "ERR_MAPS_NO_RESOURCE")          // 没有可用账号
	ErrMapsNoInterface            = status.Error(codes.ResourceExhausted, "ERR_MAPS_NO_INTERFACE")         // 没有可用接口
	ErrMapsReachDailyLimit        = status.Error(codes.ResourceExhausted, "ERR_MAPS_REACH_DAILY_QUOT")     // 地图接口当天配额已用完
	ErrMapsMapIDNotFound          = status.Error(codes.NotFound, "ERR_MAPS_MAP_ID_NOT_FOUND")              // 地图不存在
	ErrMapsInterfaceNotSupported  = status.Error(codes.Internal, "ERR_MAPS_INTERFACE_NOT_SUPPORTED")       // 地图接口尚不支持
	ErrMapsSourceNotSupported     = status.Error(codes.Internal, "ERR_MAPS_SOURCE_NOT_SUPPORTED")          // 地图源不支持
	ErrMapsInterfaceNotFound      = status.Error(codes.Internal, "ERR_MAPS_INTERFACE_NOT_FOUND")           // 地图接口不存在
	ErrMapsAlreadyExists          = status.Error(codes.AlreadyExists, "ERR_MAPS_ALREADY_EXISTS")           // 地图已存在
	ErrMapsInterfaceAlreadyExists = status.Error(codes.AlreadyExists, "ERR_MAPS_INTERFACE_ALREADY_EXISTS") // 地图接口已存在
	ErrMapsMatrixTooManyPoints    = status.Error(codes.Internal, "ERR_MAPS_MATRIX_TOO_MANY_POINTS")        // 起点和终点乘积大于100
)
