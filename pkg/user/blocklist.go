package user

import (
	"context"
	"fmt"
	"time"

	configv1 "github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/go-common/redis"
)

type BlockList interface {
	Add(ctx context.Context, target string, ttl time.Duration) error
	BatchAdd(ctx context.Context, targets []*BlockItem) error
	Exists(ctx context.Context, target string) (bool, error)
}

type SessionBlockList struct {
	rdb    *redis.Redis
	prefix string
}

func NewSessionBlockList(c *configv1.SessionBlockList, rdb *redis.Redis) BlockList {
	return &SessionBlockList{
		rdb:    rdb,
		prefix: c.Prefix,
	}
}

func (r *SessionBlockList) Add(ctx context.Context, sessionID string, ttl time.Duration) error {
	return r.rdb.Set(ctx, r.getBlockListKey(sessionID), "1", ttl).Err()
}

func (r *SessionBlockList) BatchAdd(ctx context.Context, sessions []*BlockItem) error {
	if len(sessions) == 0 {
		return nil
	}
	pipe := r.rdb.Pipeline()
	for _, session := range sessions {
		key := r.getBlockListKey(session.Target)
		pipe.Set(ctx, key, "1", session.TTL)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *SessionBlockList) Exists(ctx context.Context, sessionID string) (bool, error) {
	exists, err := r.rdb.Exists(ctx, r.getBlockListKey(sessionID)).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (r *SessionBlockList) getBlockListKey(sessionID string) string {
	return fmt.Sprintf("%s:{%s}", r.prefix, sessionID)
}
