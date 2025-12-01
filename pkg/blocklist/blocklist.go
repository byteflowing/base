package blocklist

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/byteflowing/base/pkg/redis"
	redisv9 "github.com/redis/go-redis/v9"
)

type BlockItem struct {
	Target string
	TTL    time.Duration
}

type BlockList struct {
	prefix string
	rdb    *redis.Redis
}

func NewBlockList(prefix string, rdb *redis.Redis) *BlockList {
	return &BlockList{
		prefix: prefix,
		rdb:    rdb,
	}
}

func (b *BlockList) Add(ctx context.Context, target string, ttl time.Duration) error {
	key := b.getKey(target)
	return b.rdb.Set(ctx, key, "1", ttl).Err()
}

func (b *BlockList) BatchAdd(ctx context.Context, items []*BlockItem) error {
	if len(items) == 0 {
		return nil
	}
	pipe := b.rdb.TxPipeline()
	for _, item := range items {
		key := b.getKey(item.Target)
		pipe.Set(ctx, key, "1", item.TTL)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (b *BlockList) Exists(ctx context.Context, target string) (bool, error) {
	key := b.getKey(target)
	n, err := b.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (b *BlockList) BatchExists(ctx context.Context, targets []string) ([]bool, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	count := len(targets)
	keys := make([]string, count)
	for idx, target := range targets {
		keys[idx] = b.getKey(target)
	}
	cmds := make([]*redisv9.IntCmd, count)
	_, err := b.rdb.Pipelined(ctx, func(pipe redisv9.Pipeliner) error {
		for i, key := range keys {
			cmds[i] = pipe.Exists(ctx, key)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	result := make([]bool, count)
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		result[i] = val == 1
	}
	return result, nil
}

func (b *BlockList) Remove(ctx context.Context, target string) error {
	key := b.getKey(target)
	return b.rdb.Del(ctx, key).Err()
}

func (b *BlockList) BatchRemove(ctx context.Context, targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	keys := make([]string, len(targets))
	for idx, target := range targets {
		keys[idx] = b.getKey(target)
	}
	return b.rdb.Del(ctx, keys...).Err()
}

func (b *BlockList) TTL(ctx context.Context, target string) (time.Duration, error) {
	key := b.getKey(target)
	return b.rdb.TTL(ctx, key).Result()
}

func (b *BlockList) BatchTTL(ctx context.Context, targets []string) ([]time.Duration, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	keys := make([]string, len(targets))
	for i, t := range targets {
		keys[i] = b.getKey(t)
	}
	cmds := make([]*redisv9.DurationCmd, len(keys))
	_, err := b.rdb.Pipelined(ctx, func(pipe redisv9.Pipeliner) error {
		for i, key := range keys {
			cmds[i] = pipe.TTL(ctx, key)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	results := make([]time.Duration, len(keys))
	for i, cmd := range cmds {
		d, err := cmd.Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
		results[i] = d
	}
	return results, nil
}

func (b *BlockList) getKey(target string) string {
	return fmt.Sprintf("%s:%s", b.prefix, target)
}
