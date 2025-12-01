package limiter

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	r *rate.Limiter
}

// NewRateLimiter 创建一个限流器，允许在 duration 内执行 maxRequests 次请求
// 例如：NewRateLimiter(5*time.Second, 10, 5) 表示每5秒最多10次请求，突发容量5
func NewRateLimiter(duration time.Duration, maxRequests, burst uint64) *RateLimiter {
	var rl *rate.Limiter
	if maxRequests == 0 {
		rl = rate.NewLimiter(rate.Limit(0), int(burst)) // 禁止所有请求
	} else {
		interval := duration / time.Duration(maxRequests)
		rl = rate.NewLimiter(
			rate.Every(interval),
			int(burst),
		)
	}
	return &RateLimiter{r: rl}
}

// Allow 检查是否允许请求
func (l *RateLimiter) Allow() bool {
	return l.r.Allow()
}

func (l *RateLimiter) AllowN(t time.Time, n int) bool {
	return l.r.AllowN(t, n)
}

// Wait 等待直到可以执行请求
func (l *RateLimiter) Wait(ctx context.Context) error {
	return l.r.Wait(ctx)
}

func (l *RateLimiter) WaitN(ctx context.Context, n int) error {
	return l.r.WaitN(ctx, n)
}

func (l *RateLimiter) Tokens() float64 {
	return l.r.Tokens()
}

func (l *RateLimiter) Burst() int {
	return l.r.Burst()
}

func (l *RateLimiter) SetLimit(duration time.Duration, maxRequests uint64) {
	if maxRequests == 0 {
		l.r.SetLimit(rate.Limit(0)) // 禁止所有请求
		return
	}
	interval := duration / time.Duration(maxRequests)
	newLimit := rate.Every(interval)
	l.r.SetLimit(newLimit)
}

func (l *RateLimiter) SetBurst(burst uint64) {
	l.r.SetBurst(int(burst))
}
