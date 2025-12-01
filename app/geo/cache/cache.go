package cache

import (
	"context"
	"errors"

	"github.com/byteflowing/base/app/geo/pack"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/cache"
	"github.com/byteflowing/base/pkg/redis"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
	"google.golang.org/protobuf/proto"
)

// Cache 定义了GeoService的缓存接口
type Cache interface {
	Lock(ctx context.Context, target string) (identifier string, err error)
	Unlock(ctx context.Context, target, identifier string) error
	GetPhoneKey() string
	GetCountryKey() string
	GetCountryCca2RegionKey(cca2 string) string
	CheckExists(ctx context.Context, key string) (exist bool, err error)

	GetAllPhoneCodes(ctx context.Context, req *geov1.GetAllPhoneCodesReq) (resp *geov1.GetAllPhoneCodesResp, err error)
	GetPhoneCodeByID(ctx context.Context, req *geov1.GetPhoneCodeByIdReq) (resp *geov1.GetPhoneCodeByIdResp, err error)
	SetAllPhoneCodes(ctx context.Context, codes []*geov1.GeoPhoneCode) error
	DeleteAllPhoneCodes(ctx context.Context) error

	GetAllCountries(ctx context.Context, req *geov1.GetAllCountriesReq) (resp *geov1.GetAllCountriesResp, err error)
	GetCountryByCca2(ctx context.Context, req *geov1.GetCountryByCca2Req) (resp *geov1.GetCountryByCca2Resp, err error)
	GetCountryByID(ctx context.Context, req *geov1.GetCountryByIdReq) (resp *geov1.GetCountryByIdResp, err error)
	SetAllCountries(ctx context.Context, countries []*geov1.GeoCountry) error
	DeleteAllCountries(ctx context.Context) error

	GetRegionsByCountryCca2(ctx context.Context, countryCca2 string) ([]*geov1.GeoRegion, error)
	SetRegionsByCountryCca2(ctx context.Context, countryCca2 string, regions []*geov1.GeoRegion) error
	DeleteRegionsByCountryCca2(ctx context.Context, countryCca2 string) error
}

func New(
	cfg *geov1.CacheConfig,
	rdb *redis.Redis,
	localCache *cache.Cache,
) Cache {
	if cfg.Type == geov1.CacheConfig_CACHE_TYPE_LOCAL {
		return NewLocalCache(cfg, localCache)
	} else if cfg.Type == geov1.CacheConfig_CACHE_TYPE_REDIS {
		return NewRedisCache(cfg, rdb)
	}
	panic("invalid cache type")
}

//----------------------------------------------------------------------------------------------------------------------
// 公共实现

func IsCacheNotFoundErr(err error) bool {
	return errors.Is(err, ecode.ErrCacheNotExist)
}

func getAllPhoneCodes(data []byte, lang enumv1.Language) (resp *geov1.GetAllPhoneCodesResp, err error) {
	phoneCodes := &geov1.AllPhoneCodes{}
	if err := proto.Unmarshal(data, phoneCodes); err != nil {
		return nil, err
	}
	for _, phoneCode := range phoneCodes.Codes {
		phoneCode.Name = pack.GetLangName(phoneCode.MultiLang, lang)
		phoneCode.MultiLang = nil
	}
	return &geov1.GetAllPhoneCodesResp{PhoneCodes: phoneCodes.Codes}, nil
}

func getPhoneCodeById(data []byte, id int64, lang enumv1.Language) (resp *geov1.GetPhoneCodeByIdResp, err error) {
	phoneCodes := &geov1.AllPhoneCodes{}
	if err := proto.Unmarshal(data, phoneCodes); err != nil {
		return nil, err
	}
	for _, phoneCode := range phoneCodes.Codes {
		if phoneCode.Id == id {
			phoneCode.Name = pack.GetLangName(phoneCode.MultiLang, lang)
			phoneCode.MultiLang = nil
			return &geov1.GetPhoneCodeByIdResp{PhoneCode: phoneCode}, nil
		}
	}
	return nil, ecode.ErrPhoneCodeNotExist
}

func getAllCountries(data []byte, lang enumv1.Language) (resp *geov1.GetAllCountriesResp, err error) {
	countryCodes := &geov1.AllCountryCodes{}
	if err := proto.Unmarshal(data, countryCodes); err != nil {
		return nil, err
	}
	for _, countryCode := range countryCodes.Codes {
		countryCode.Name = pack.GetLangName(countryCode.MultiLang, lang)
		countryCode.MultiLang = nil
	}
	return &geov1.GetAllCountriesResp{Countries: countryCodes.Codes}, nil
}

func getCountryByCca2(data []byte, cca2 string, lang enumv1.Language) (resp *geov1.GetCountryByCca2Resp, err error) {
	countryCodes := &geov1.AllCountryCodes{}
	if err := proto.Unmarshal(data, countryCodes); err != nil {
		return nil, err
	}
	for _, countryCode := range countryCodes.Codes {
		if countryCode.Cca2 == cca2 {
			countryCode.Name = pack.GetLangName(countryCode.MultiLang, lang)
			countryCode.MultiLang = nil
			return &geov1.GetCountryByCca2Resp{Country: countryCode}, nil
		}
	}
	return nil, ecode.ErrCountryCodeNotExist
}

func getCountryByID(data []byte, id int64, lang enumv1.Language) (resp *geov1.GetCountryByIdResp, err error) {
	countryCodes := &geov1.AllCountryCodes{}
	if err := proto.Unmarshal(data, countryCodes); err != nil {
		return nil, err
	}
	for _, countryCode := range countryCodes.Codes {
		if countryCode.Id == id {
			countryCode.Name = pack.GetLangName(countryCode.MultiLang, lang)
			countryCode.MultiLang = nil
			return &geov1.GetCountryByIdResp{Country: countryCode}, nil
		}
	}
	return nil, ecode.ErrCountryCodeNotExist
}

func getRegionsByCountryCca2(data []byte) ([]*geov1.GeoRegion, error) {
	regions := &geov1.CountryRegions{}
	if err := proto.Unmarshal(data, regions); err != nil {
		return nil, err
	}
	return regions.Regions, nil
}
