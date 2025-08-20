package user

import (
	"context"
	"fmt"

	"github.com/byteflowing/base/ecode"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	"github.com/byteflowing/go-common/redis"
)

type Limiter interface {
	Allow(ctx context.Context, uid int64, t enumsv1.UserLimitType) error
}

type RedisLimiter struct {
	rdb    *redis.Redis
	prefix string
	rules  map[enumsv1.UserLimitType]*commonv1.LimitRule
}

func (r *RedisLimiter) Allow(ctx context.Context, uid int64, t enumsv1.UserLimitType) error {
	key := fmt.Sprintf("%s:%d", r.prefix, t.Number())
	duration := r.rules[t].Duration.AsDuration()
	limit := r.rules[t].Limit
	ok, err := r.rdb.AllowFixedLimit(ctx, key, duration, uint32(limit))
	if err != nil {
		return err
	}
	if !ok {
		return ecode.ErrTooManyRequests
	}
	return nil
}
