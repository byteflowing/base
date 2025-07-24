package limiter

import (
	"context"

	"github.com/byteflowing/go-common/redis"
)

type Limiter interface {
	Allow(ctx context.Context, target string) (ok bool, window *redis.Window, err error)
}

// SlidingLimiter 用户限制短信发送频率
// 可以通过windows参数设置限制的时间窗口
// 比如 1分钟 5分钟 10分钟
type SlidingLimiter struct {
	limiter *redis.Limiter
}

func NewSlidingLimiter(rdb *redis.Redis, prefix string, windows []*redis.Window) Limiter {
	limiter := redis.NewLimiter(rdb, prefix, windows)
	return &SlidingLimiter{limiter: limiter}
}

func (s *SlidingLimiter) Allow(ctx context.Context, target string) (ok bool, window *redis.Window, err error) {
	return s.limiter.Allow(ctx, target)
}

// DailyLimiter 用于限制一个自然天内的请求
// Allow返回的window恒为nil
type DailyLimiter struct {
	rdb    *redis.Redis
	prefix string
	max    uint32
}

func NewDailyLimiter(rdb *redis.Redis, prefix string, maxCount uint32) Limiter {
	return &DailyLimiter{
		rdb:    rdb,
		prefix: prefix,
		max:    maxCount,
	}
}

func (f *DailyLimiter) Allow(ctx context.Context, target string) (ok bool, window *redis.Window, err error) {
	ok, err = f.rdb.AllowDailyLimit(ctx, f.prefix, target, f.max)
	return ok, nil, err
}
