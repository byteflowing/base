package user

import (
	"context"
	"fmt"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/go-common/ratelimit"
	"github.com/byteflowing/go-common/redis"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	limiterv1 "github.com/byteflowing/proto/gen/go/limiter/v1"
	userv1 "github.com/byteflowing/proto/gen/go/services/user/v1"
)

type AuthLimiter struct {
	rdb        *redis.Redis
	prefix     string
	errPrefix  string
	rules      map[enumsv1.UserAuthLimitType]*limiterv1.LimitRule
	errLimiter *ratelimit.RedisLimiter
}

func NewAuthLimiter(c *userv1.UserAuthLimiter, rdb *redis.Redis) *AuthLimiter {
	errLimiter := ratelimit.NewRedisLimiter(rdb, c.ErrPrefix, c.SignInErrRules)
	var rules = make(map[enumsv1.UserAuthLimitType]*limiterv1.LimitRule, 2)
	rules[enumsv1.UserAuthLimitType_USER_AUTH_LIMIT_TYPE_SIGN_IN] = c.SignInRule
	rules[enumsv1.UserAuthLimitType_USER_AUTH_LIMIT_TYPE_REFRESH] = c.RefreshRule
	return &AuthLimiter{
		rdb:        rdb,
		prefix:     c.Prefix,
		errPrefix:  c.ErrPrefix,
		rules:      rules,
		errLimiter: errLimiter,
	}
}

func (r *AuthLimiter) Allow(ctx context.Context, uid int64, t enumsv1.UserAuthLimitType) error {
	key := fmt.Sprintf("%s:%v:%d", r.prefix, uid, t.Number())
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

func (r *AuthLimiter) AllowErr(ctx context.Context, uid int64) (rule *limiterv1.LimitRule, allow bool, err error) {
	key := r.getErrKey(uid)
	allow, rule, err = r.errLimiter.Allow(ctx, key)
	if err != nil {
		return nil, false, err
	}
	if !allow {
		return rule, false, nil
	}
	return nil, true, nil
}

func (r *AuthLimiter) ResetErr(ctx context.Context, uid int64) error {
	key := r.getErrKey(uid)
	return r.errLimiter.Reset(ctx, key)
}

func (r *AuthLimiter) getErrKey(uid int64) string {
	return fmt.Sprintf("%s:%d", r.errPrefix, uid)
}

func (r *AuthLimiter) getError(t enumsv1.UserAuthLimitType) error {
	if t == enumsv1.UserAuthLimitType_USER_AUTH_LIMIT_TYPE_SIGN_IN {
		return ecode.ErrUserSignInTooMany
	} else if t == enumsv1.UserAuthLimitType_USER_AUTH_LIMIT_TYPE_REFRESH {
		return ecode.ErrUserRefreshTooMany
	}
	return ecode.ErrTooManyRequests
}
