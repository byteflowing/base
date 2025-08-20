package common

import (
	"context"
	"fmt"

	configv1 "github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/go-common/redis"
)

type DistributedLock interface {
	Lock(ctx context.Context, target string) (identifier string, err error)
	Unlock(ctx context.Context, target, identifier string) error
	ReNew(ctx context.Context, target, identifier string) error
}

func NewDistributedLock(rdb *redis.Redis, c *configv1.DistributedLock) DistributedLock {
	return &RedisDistributedLock{
		rdb:    rdb,
		config: c,
	}
}

type RedisDistributedLock struct {
	rdb    *redis.Redis
	config *configv1.DistributedLock
}

func (r *RedisDistributedLock) Lock(ctx context.Context, target string) (identifier string, err error) {
	return r.rdb.Lock(ctx, r.getKey(target), r.config.Ttl.AsDuration())
}

func (r *RedisDistributedLock) Unlock(ctx context.Context, target, identifier string) error {
	return r.rdb.Unlock(ctx, r.getKey(target), identifier)
}

func (r *RedisDistributedLock) ReNew(ctx context.Context, target, identifier string) error {
	return r.rdb.RenewLock(ctx, r.getKey(target), identifier, r.config.Ttl.AsDuration())
}

func (r *RedisDistributedLock) getKey(target string) string {
	return fmt.Sprintf("%s:%s", r.config.Prefix, target)
}
