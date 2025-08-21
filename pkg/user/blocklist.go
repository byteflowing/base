package user

import (
	"context"
	"fmt"
	"time"

	configv1 "github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/go-common/redis"
)

type BlockList interface {
	Add(ctx context.Context, sessionID string, ttl time.Duration) error
	BatchAdd(ctx context.Context, sessions []*SessionItem) error
	Exists(ctx context.Context, sessionID string) (bool, error)
}

type RedisBlockList struct {
	rdb    *redis.Redis
	prefix string
}

func NewRedisBlockList(rdb *redis.Redis, c *configv1.UserBlockList) *RedisBlockList {
	return &RedisBlockList{
		rdb:    rdb,
		prefix: c.Prefix,
	}
}

func (r *RedisBlockList) Add(ctx context.Context, sessionID string, ttl time.Duration) error {
	return r.rdb.Set(ctx, r.getBlockListKey(sessionID), "1", ttl).Err()
}

func (r *RedisBlockList) BatchAdd(ctx context.Context, sessions []*SessionItem) error {
	if len(sessions) == 0 {
		return nil
	}
	pipe := r.rdb.Pipeline()
	for _, session := range sessions {
		key := r.getBlockListKey(session.SessionID)
		pipe.Set(ctx, key, "1", session.TTL)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisBlockList) Exists(ctx context.Context, sessionID string) (bool, error) {
	exists, err := r.rdb.Exists(ctx, r.getBlockListKey(sessionID)).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (r *RedisBlockList) getBlockListKey(sessionID string) string {
	return fmt.Sprintf("%s:{%s}", r.prefix, sessionID)
}
