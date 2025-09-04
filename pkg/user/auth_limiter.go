package user

import (
	"context"
	"fmt"

	"github.com/byteflowing/base/ecode"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	configv1 "github.com/byteflowing/base/gen/config/v1"
	enumsv1 "github.com/byteflowing/base/gen/enums/v1"
	"github.com/byteflowing/base/pkg/common"
	"github.com/byteflowing/go-common/redis"
	"github.com/byteflowing/go-common/trans"
	"google.golang.org/protobuf/types/known/durationpb"
)

type AuthLimiter struct {
	rdb        *redis.Redis
	prefix     string
	errPrefix  string
	rules      map[enumsv1.UserAuthLimitType]*commonv1.LimitRule
	errLimiter *redis.Limiter
}

func NewAuthLimiter(c *configv1.UserAuthLimiter, rdb *redis.Redis) *AuthLimiter {
	errRules := common.ConvertLimitsToWindows(c.SignInErrRules)
	errLimiter := redis.NewLimiter(rdb, c.ErrPrefix, errRules)
	var rules = make(map[enumsv1.UserAuthLimitType]*commonv1.LimitRule, 2)
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

func (r *AuthLimiter) AllowErr(ctx context.Context, uid int64) (rule *commonv1.LimitRule, allow bool, err error) {
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
