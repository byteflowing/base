package cache

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/cache"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
)

// LocalCache 实现了本地缓存
type LocalCache struct {
	cfg   *geov1.CacheConfig
	cache *cache.Cache
	lock  *cache.Lock
}

// NewLocalCache 创建一个新的本地缓存实例
func NewLocalCache(cfg *geov1.CacheConfig, localCache *cache.Cache) *LocalCache {
	if localCache == nil {
		panic("local cache is nil")
	}
	lock := cache.NewLock(localCache, &cache.LockOption{
		Prefix: cfg.Lock.Prefix,
		Tries:  int(cfg.Lock.Tries),
		TTL:    cfg.Lock.Ttl.AsDuration(),
		Wait:   cfg.Lock.Wait.AsDuration(),
	})
	return &LocalCache{
		cfg:   cfg,
		cache: localCache,
		lock:  lock,
	}
}

func (c *LocalCache) Lock(ctx context.Context, target string) (identifier string, err error) {
	key := c.getLockKey(target)
	return c.lock.Acquire(ctx, key)
}

func (c *LocalCache) Unlock(ctx context.Context, target, identifier string) error {
	key := c.getLockKey(target)
	return c.lock.Release(ctx, key, identifier)
}

func (c *LocalCache) GetPhoneKey() string {
	return c.cfg.PhoneCodeCache.Prefix
}

func (c *LocalCache) GetCountryKey() string {
	return c.cfg.CountryCodeCache.Prefix
}

func (c *LocalCache) GetCountryCca2RegionKey(cca2 string) string {
	return c.cfg.RegionCodeCache.Prefix + ":" + cca2
}

func (c *LocalCache) CheckExists(_ context.Context, key string) (exist bool, err error) {
	return c.cache.Exists(key)
}

func (c *LocalCache) GetAllPhoneCodes(_ context.Context, req *geov1.GetAllPhoneCodesReq) (resp *geov1.GetAllPhoneCodesResp, err error) {
	key := c.GetPhoneKey()
	data, err := c.cache.Get(key)
	if err := c.parseErr(err); err != nil {
		return nil, err
	}
	return getAllPhoneCodes(data, req.Lang)
}

func (c *LocalCache) GetPhoneCodeByID(_ context.Context, req *geov1.GetPhoneCodeByIdReq) (resp *geov1.GetPhoneCodeByIdResp, err error) {
	key := c.GetPhoneKey()
	data, err := c.cache.Get(key)
	if err := c.parseErr(err); err != nil {
		return nil, err
	}
	return getPhoneCodeById(data, req.Id, req.Lang)
}

func (c *LocalCache) SetAllPhoneCodes(_ context.Context, phoneCodes []*geov1.GeoPhoneCode) error {
	key := c.GetPhoneKey()
	codes := &geov1.AllPhoneCodes{Codes: phoneCodes}
	data, err := proto.Marshal(codes)
	if err != nil {
		return err
	}
	return c.cache.Set(key, data, int(c.cfg.PhoneCodeCache.Duration.Seconds))
}

func (c *LocalCache) DeleteAllPhoneCodes(_ context.Context) error {
	key := c.GetPhoneKey()
	c.cache.Delete(key)
	return nil
}

func (c *LocalCache) GetAllCountries(_ context.Context, req *geov1.GetAllCountriesReq) (resp *geov1.GetAllCountriesResp, err error) {
	key := c.GetCountryKey()
	data, err := c.cache.Get(key)
	if err := c.parseErr(err); err != nil {
		return nil, err
	}
	return getAllCountries(data, req.Lang)
}

func (c *LocalCache) GetCountryByCca2(_ context.Context, req *geov1.GetCountryByCca2Req) (resp *geov1.GetCountryByCca2Resp, err error) {
	key := c.GetCountryKey()
	data, err := c.cache.Get(key)
	if err := c.parseErr(err); err != nil {
		return nil, err
	}
	return getCountryByCca2(data, req.Cca2, req.Lang)
}

func (c *LocalCache) GetCountryByID(_ context.Context, req *geov1.GetCountryByIdReq) (resp *geov1.GetCountryByIdResp, err error) {
	key := c.GetCountryKey()
	data, err := c.cache.Get(key)
	if err := c.parseErr(err); err != nil {
		return nil, err
	}
	return getCountryByID(data, req.Id, req.Lang)
}

func (c *LocalCache) SetAllCountries(_ context.Context, countries []*geov1.GeoCountry) error {
	key := c.GetCountryKey()
	countryCodes := &geov1.AllCountryCodes{Codes: countries}
	data, err := proto.Marshal(countryCodes)
	if err != nil {
		return err
	}
	return c.cache.Set(key, data, int(c.cfg.CountryCodeCache.Duration.Seconds))
}

func (c *LocalCache) DeleteAllCountries(_ context.Context) error {
	key := c.GetCountryKey()
	c.cache.Delete(key)
	return nil
}

// GetRegionsByCountryCca2 获取指定国家的三级region信息
// 这里没有转换name到对应语言，需要根据使用在业务层转换
func (c *LocalCache) GetRegionsByCountryCca2(_ context.Context, countryCca2 string) ([]*geov1.GeoRegion, error) {
	key := c.GetCountryCca2RegionKey(countryCca2)
	data, err := c.cache.Get(key)
	if err := c.parseErr(err); err != nil {
		return nil, err
	}
	return getRegionsByCountryCca2(data)
}

func (c *LocalCache) SetRegionsByCountryCca2(_ context.Context, countryCca2 string, regions []*geov1.GeoRegion) error {
	key := c.GetCountryCca2RegionKey(countryCca2)
	rs := &geov1.CountryRegions{Regions: regions}
	data, err := proto.Marshal(rs)
	if err != nil {
		return err
	}
	return c.cache.Set(key, data, int(c.cfg.RegionCodeCache.Duration.Seconds))
}

func (c *LocalCache) DeleteRegionsByCountryCca2(_ context.Context, countryCca2 string) error {
	key := c.GetCountryCca2RegionKey(countryCca2)
	c.cache.Delete(key)
	return nil
}

func (c *LocalCache) parseErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, cache.ErrNotFound) {
		return ecode.ErrCacheNotExist
	}
	return err
}

func (c *LocalCache) getLockKey(target string) string {
	return c.cfg.Lock.Prefix + ":" + target
}
