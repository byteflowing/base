package cache

import (
	"context"
	"errors"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/redis"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
	"google.golang.org/protobuf/proto"
)

// RedisCache 实现了Redis缓存
type RedisCache struct {
	cfg *geov1.CacheConfig

	rdb  *redis.Redis
	lock *redis.Lock
}

func NewRedisCache(cfg *geov1.CacheConfig, rdb *redis.Redis) *RedisCache {
	if rdb == nil {
		panic("rdb is nil")
	}
	lock := redis.NewLock(rdb, &redis.LockOption{
		Prefix: cfg.Lock.Prefix,
		Tries:  int(cfg.Lock.Tries),
		TTL:    cfg.Lock.Ttl.AsDuration(),
		Wait:   cfg.Lock.Wait.AsDuration(),
	})
	return &RedisCache{
		cfg:  cfg,
		rdb:  rdb,
		lock: lock,
	}
}

func (r *RedisCache) Lock(ctx context.Context, target string) (identifier string, err error) {
	key := r.getLockKey(target)
	return r.lock.Acquire(ctx, key)
}

func (r *RedisCache) Unlock(ctx context.Context, target, identifier string) error {
	key := r.getLockKey(target)
	return r.lock.Release(ctx, key, identifier)
}

func (r *RedisCache) GetPhoneKey() string {
	return r.cfg.PhoneCodeCache.Prefix
}

func (r *RedisCache) GetCountryKey() string {
	return r.cfg.CountryCodeCache.Prefix
}

func (r *RedisCache) GetCountryCca2RegionKey(cca2 string) string {
	return r.cfg.RegionCodeCache.Prefix + ":" + cca2
}

func (r *RedisCache) CheckExists(ctx context.Context, key string) (exist bool, err error) {
	data, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return data > 0, nil
}

func (r *RedisCache) GetAllPhoneCodes(ctx context.Context, req *geov1.GetAllPhoneCodesReq) (resp *geov1.GetAllPhoneCodesResp, err error) {
	key := r.GetPhoneKey()
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err := r.parseErr(err); err != nil {
		return nil, err
	}
	return getAllPhoneCodes(data, req.Lang)
}

func (r *RedisCache) GetPhoneCodeByID(ctx context.Context, req *geov1.GetPhoneCodeByIdReq) (resp *geov1.GetPhoneCodeByIdResp, err error) {
	key := r.GetPhoneKey()
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err := r.parseErr(err); err != nil {
		return nil, err
	}
	return getPhoneCodeById(data, req.Id, req.Lang)
}

func (r *RedisCache) SetAllPhoneCodes(ctx context.Context, codes []*geov1.GeoPhoneCode) error {
	key := r.GetPhoneKey()
	phoneCodes := &geov1.AllPhoneCodes{Codes: codes}
	data, err := proto.Marshal(phoneCodes)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, r.cfg.PhoneCodeCache.Duration.AsDuration()).Err()
}

func (r *RedisCache) DeleteAllPhoneCodes(ctx context.Context) error {
	key := r.GetPhoneKey()
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisCache) GetAllCountries(ctx context.Context, req *geov1.GetAllCountriesReq) (resp *geov1.GetAllCountriesResp, err error) {
	key := r.GetCountryKey()
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err := r.parseErr(err); err != nil {
		return nil, err
	}
	return getAllCountries(data, req.Lang)
}

func (r *RedisCache) GetCountryByCca2(ctx context.Context, req *geov1.GetCountryByCca2Req) (resp *geov1.GetCountryByCca2Resp, err error) {
	key := r.GetCountryKey()
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err := r.parseErr(err); err != nil {
		return nil, err
	}
	return getCountryByCca2(data, req.Cca2, req.Lang)
}

func (r *RedisCache) GetCountryByID(ctx context.Context, req *geov1.GetCountryByIdReq) (resp *geov1.GetCountryByIdResp, err error) {
	key := r.GetCountryKey()
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err := r.parseErr(err); err != nil {
		return nil, err
	}
	return getCountryByID(data, req.Id, req.Lang)
}

func (r *RedisCache) SetAllCountries(ctx context.Context, countries []*geov1.GeoCountry) error {
	key := r.GetCountryKey()
	countryCodes := &geov1.AllCountryCodes{Codes: countries}
	data, err := proto.Marshal(countryCodes)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, r.cfg.CountryCodeCache.Duration.AsDuration()).Err()
}

func (r *RedisCache) DeleteAllCountries(ctx context.Context) error {
	key := r.GetCountryKey()
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisCache) GetRegionsByCountryCca2(ctx context.Context, countryCca2 string) ([]*geov1.GeoRegion, error) {
	key := r.GetCountryCca2RegionKey(countryCca2)
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err := r.parseErr(err); err != nil {
		return nil, err
	}
	return getRegionsByCountryCca2(data)
}

func (r *RedisCache) SetRegionsByCountryCca2(ctx context.Context, countryCca2 string, regions []*geov1.GeoRegion) error {
	key := r.GetCountryCca2RegionKey(countryCca2)
	countryRegions := &geov1.CountryRegions{Regions: regions}
	data, err := proto.Marshal(countryRegions)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, data, r.cfg.RegionCodeCache.Duration.AsDuration()).Err()
}

func (r *RedisCache) DeleteRegionsByCountryCca2(ctx context.Context, countryCca2 string) error {
	key := r.GetCountryCca2RegionKey(countryCca2)
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisCache) parseErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, redis.Nil) {
		return ecode.ErrCacheNotExist
	}
	return err
}

func (r *RedisCache) getLockKey(target string) string {
	return r.cfg.Lock.Prefix + ":" + target
}
