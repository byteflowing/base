package limiter

import (
	"context"
	_ "embed"
	"errors"
	"strconv"
	"time"

	"github.com/byteflowing/base/pkg/redis"
)

//go:embed script_distributed_limiter.lua
var scriptDistributedRateLimiter string

type DistributedRateLimiter struct {
	rdb *redis.Redis
}

func NewDistributedRateLimiter(rdb *redis.Redis) *DistributedRateLimiter {
	return &DistributedRateLimiter{rdb: rdb}
}

type DistributedRateLimiterResult struct {
	Allowed   bool
	Remaining float64
	Wait      time.Duration
}

// Allow 判断是否允许通过
func (d *DistributedRateLimiter) Allow(
	ctx context.Context,
	key string,
	capacity int64,
	interval time.Duration,
	requested int64,
) (*DistributedRateLimiterResult, error) {
	res, err := d.rdb.Eval(
		ctx,
		scriptDistributedRateLimiter,
		[]string{key},
		capacity,
		interval.Microseconds(),
		requested,
	).Result()
	if err != nil {
		return nil, err
	}
	arr, ok := res.([]interface{})
	if !ok || len(arr) != 3 {
		return nil, errors.New("invalid redis lua response")
	}
	allowed := arr[0].(int64) == 1
	remaining := parseFloat(arr[1])
	wait := time.Duration(parseFloat(arr[2])) * time.Millisecond

	return &DistributedRateLimiterResult{
		Allowed:   allowed,
		Remaining: remaining,
		Wait:      wait,
	}, nil
}

// Wait 等待直到允许通过（阻塞直到可用 token）
func (d *DistributedRateLimiter) Wait(
	ctx context.Context,
	key string,
	capacity int64,
	interval time.Duration,
	requested int64,
) error {
	for {
		r, err := d.Allow(ctx, key, capacity, interval, requested)
		if err != nil {
			return err
		}
		if r.Allowed {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.Wait):
		}
	}
}

func parseFloat(v interface{}) float64 {
	switch x := v.(type) {
	case int64:
		return float64(x)
	case float64:
		return x
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	default:
		return 0
	}
}
