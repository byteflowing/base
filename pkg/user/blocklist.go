package user

import (
	"context"
	"fmt"
	"time"

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

func NewRedisBlockList(rdb *redis.Redis, prefix string) *RedisBlockList {
	return &RedisBlockList{
		rdb:    rdb,
		prefix: prefix,
	}
}

func (r *RedisBlockList) Add(ctx context.Context, sessionID string, ttl time.Duration) error {
	return r.rdb.Set(ctx, r.getKey(sessionID), "1", ttl).Err()
}

func (r *RedisBlockList) BatchAdd(ctx context.Context, sessions []*SessionItem) error {
	if len(sessions) == 0 {
		return nil
	}
	pipe := r.rdb.Pipeline()
	for _, session := range sessions {
		key := r.getKey(session.SessionID)
		pipe.Set(ctx, key, "1", session.TTL)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisBlockList) Exists(ctx context.Context, sessionID string) (bool, error) {
	exists, err := r.rdb.Exists(ctx, r.getKey(sessionID)).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (r *RedisBlockList) getKey(sessionID string) string {
	return fmt.Sprintf("%s:{%s}", r.prefix, sessionID)
}
