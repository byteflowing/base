package quota

import (
	"context"
	_ "embed"
	"time"

	"github.com/byteflowing/base/pkg/redis"
)

type FixedQuota struct {
	rdb *redis.Redis
}

func NewFixedQuota(rdb *redis.Redis) *FixedQuota {
	return &FixedQuota{rdb: rdb}
}

// Allow 检查是否允许消耗一个配额
// - key: 限流键
// - ttl: 窗口周期
// - quota: 配额上限
func (f *FixedQuota) Allow(
	ctx context.Context,
	key string,
	ttl time.Duration,
	quota int,
) (*Result, error) {
	return f.allow(ctx, key, ttl, quota, 1)
}

// AllowN 检查是否允许消耗一个配额
// - key: 限流键
// - ttl: 窗口周期
// - quota: 配额上限
// - n: 一次允许n次
func (f *FixedQuota) AllowN(
	ctx context.Context,
	key string,
	ttl time.Duration,
	quota, n int,
) (*Result, error) {
	return f.allow(ctx, key, ttl, quota, n)
}

func (f *FixedQuota) Decr(ctx context.Context, key string) (newValue int64, err error) {
	return f.decr(ctx, key, 1)
}

func (f *FixedQuota) DecrN(ctx context.Context, key string, n int) (newValue int64, err error) {
	return f.decr(ctx, key, n)
}

func (f *FixedQuota) Reset(ctx context.Context, key string) (err error) {
	return f.rdb.Del(ctx, key).Err()
}

func (f *FixedQuota) GetDetail(ctx context.Context, key string) (*Detail, error) {
	return getDetail(ctx, key, f.rdb)
}

func (f *FixedQuota) allow(
	ctx context.Context,
	key string,
	ttl time.Duration,
	quota, n int,
) (*Result, error) {
	res, err := f.rdb.Eval(ctx, scriptQuota, []string{key}, ttl.Milliseconds(), quota, n).Result()
	if err != nil {
		return nil, err
	}
	result, err := parseResult(res, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (f *FixedQuota) decr(ctx context.Context, key string, n int) (newValue int64, err error) {
	return f.rdb.Eval(ctx, scriptDecr, []string{key}, n).Int64()
}
