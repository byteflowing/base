package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/go-common/redis"
)

type BlockList interface {
	Add(ctx context.Context, sessionID string, ttl time.Duration) error
	BatchAdd(ctx context.Context, sessions []*SessionItem) error
	Exists(ctx context.Context, sessionID string) (bool, error)
	CheckTokenLimit(ctx context.Context, uid uint64, t LimitType) error
}

type LimitType string

const (
	LimitTypeSignIn  LimitType = "signIn"
	LimitTypeRefresh LimitType = "refresh"
)

type RedisBlockList struct {
	rdb             *redis.Redis
	blockListPrefix string
	signInLimit     *LimitRule
	refreshLimit    *LimitRule
}

func NewRedisBlockList(rdb *redis.Redis, c *BlockListConfig) *RedisBlockList {
	return &RedisBlockList{
		rdb:             rdb,
		blockListPrefix: c.BlockListPrefix,
		signInLimit:     c.SignInLimit,
		refreshLimit:    c.RefreshLimit,
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
	return fmt.Sprintf("%s:{%s}", r.blockListPrefix, sessionID)
}

func (r *RedisBlockList) CheckTokenLimit(ctx context.Context, uid uint64, t LimitType) error {
	var key string
	var duration time.Duration
	var limit uint32
	switch t {
	case LimitTypeSignIn:
		key = fmt.Sprintf("%s:%d", r.signInLimit.Prefix, uid)
		duration = time.Duration(r.signInLimit.Duration) * time.Second
		limit = r.signInLimit.Limit
	case LimitTypeRefresh:
		key = fmt.Sprintf("%s:%d", r.refreshLimit.Prefix, uid)
		duration = time.Duration(r.refreshLimit.Duration) * time.Second
		limit = r.refreshLimit.Limit
	default:
		return errors.New("invalid limit type")
	}
	ok, err := r.rdb.AllowFixedLimit(ctx, key, duration, limit)
	if err != nil {
		return err
	}
	if !ok {
		return ecode.ErrTooManyRequests
	}
	return nil
}
