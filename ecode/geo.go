package ecode

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCacheNotExist             = status.Error(codes.NotFound, "ERR_CACHE_NOT_EXIST")              // 缓存不存在
	ErrCountryCodeNotExist       = status.Error(codes.NotFound, "ERR_COUNTRY_CODE_NOT_EXIST")       // 国家代码不存在
	ErrPhoneCodeNotExist         = status.Error(codes.NotFound, "ERR_PHONE_NOT_EXIST")              // 手机代码不存在
	ErrPhoneCodeNotImported      = status.Error(codes.NotFound, "ERR_PHONE_NOT_IMPORTED")           // 手机代码未导入
	ErrCountryCodeNotImported    = status.Error(codes.NotFound, "ERR_COUNTRY_CODE_NOT_IMPORTED")    // 国家代码未导入
	ErrCountryRegionsNotImported = status.Error(codes.NotFound, "ERR_COUNTRY_REGIONS_NOT_IMPORTED") // 地区信息未导入
	ErrRegionCodeNotExists       = status.Error(codes.NotFound, "ERR_REGION_CODE_NOT_EXISTS")       // 地区代码不存在
	ErrCountryCca2NotExist       = status.Error(codes.NotFound, "ERR_COUNTRY_CCA2_NOT_EXIST")       // 国家cca2编码不存在
	ErrCountryCodeExists         = status.Error(codes.AlreadyExists, "ERR_COUNTRY_CODE_EXISTS")     // 国家代码已存在
	ErrRegionCodeExists          = status.Error(codes.AlreadyExists, "ERR_REGION_CODE_EXISTS")      // 地区代码已存在
)
