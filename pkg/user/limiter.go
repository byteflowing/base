package user

import (
	"context"
	"fmt"

	"github.com/byteflowing/base/ecode"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	"github.com/byteflowing/go-common/redis"
	"github.com/byteflowing/go-common/trans"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Limiter interface {
	Allow(ctx context.Context, uid int64, t enumsv1.UserLimitType) error
	AllowErr(ctx context.Context, uid int64) (rule *commonv1.LimitRule, allow bool, err error)
	ResetErr(ctx context.Context, uid int64) error
}

type RedisLimiter struct {
	rdb        *redis.Redis
	prefix     string
	errPrefix  string
	rules      map[enumsv1.UserLimitType]*commonv1.LimitRule
	errLimiter *redis.Limiter
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
		return r.getError(t)
	}
	return nil
}

func (r *RedisLimiter) AllowErr(ctx context.Context, uid int64) (rule *commonv1.LimitRule, allow bool, err error) {
	key := r.getErrKey(uid)
	allow, window, after, err := r.errLimiter.Allow(ctx, key)
	if err != nil {
		return nil, false, err
	}
	if !allow {
		rule = &commonv1.LimitRule{
			Duration:   durationpb.New(window.Duration),
			Limit:      int32(window.Limit),
			Tag:        window.Tag,
			RetryAfter: trans.Ref(after),
		}
		return rule, false, nil
	}
	return nil, true, nil
}

func (r *RedisLimiter) ResetErr(ctx context.Context, uid int64) error {
	key := r.getErrKey(uid)
	return r.errLimiter.Reset(ctx, key)
}

func (r *RedisLimiter) getErrKey(uid int64) string {
	return fmt.Sprintf("%s:%d", r.errPrefix, uid)
}

func (r *RedisLimiter) getError(t enumsv1.UserLimitType) error {
	if t == enumsv1.UserLimitType_USER_LIMIT_TYPE_SIGN_IN {
		return ecode.ErrUserSignInTooMany
	} else if t == enumsv1.UserLimitType_USER_LIMIT_TYPE_REFRESH {
		return ecode.ErrUserRefreshTooMany
	}
	return ecode.ErrTooManyRequests
}
