package common

import (
	"context"
	"fmt"
	"time"

	"github.com/byteflowing/go-common/redis"
)

type DistributedLockConfig struct {
	TTL    uint32 // 锁的过期时间,单位s
	Prefix string // key前缀
}

type DistributedLock interface {
	Lock(ctx context.Context, target string) (identifier string, err error)
	Unlock(ctx context.Context, target, identifier string) error
	ReNew(ctx context.Context, target, identifier string) error
}

type RedisDistributedLock struct {
	rdb    *redis.Redis
	config *DistributedLockConfig
}

func (r *RedisDistributedLock) Lock(ctx context.Context, target string) (identifier string, err error) {
	return r.rdb.Lock(ctx, r.getKey(target), time.Duration(r.config.TTL)*time.Second)
}

func (r *RedisDistributedLock) Unlock(ctx context.Context, target, identifier string) error {
	return r.rdb.Unlock(ctx, r.getKey(target), identifier)
}

func (r *RedisDistributedLock) ReNew(ctx context.Context, target, identifier string) error {
	return r.rdb.RenewLock(ctx, r.getKey(target), identifier, time.Duration(r.config.TTL)*time.Second)
}

func (r *RedisDistributedLock) getKey(target string) string {
	return fmt.Sprintf("%s:%s", r.config.Prefix, target)
}
