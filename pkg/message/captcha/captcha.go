package captcha

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/go-common/idx"
	"github.com/byteflowing/go-common/redis"
)

type Captcha interface {
	Save(ctx context.Context, target, val string, fn func() error) (token string, rule *LimitRule, err error)
	Verify(ctx context.Context, token, captcha string) (ok bool, err error)
}

type Impl struct {
	config  *Config
	rdb     *redis.Redis
	limiter *redis.Limiter
}

func New(rdb *redis.Redis, c *Config) Captcha {
	limiter := redis.NewLimiter(rdb, c.KeyPrefix, c.ToWindows())
	return &Impl{
		config:  c,
		rdb:     rdb,
		limiter: limiter,
	}
}

func (i *Impl) Save(ctx context.Context, target, val string, fn func() error) (token string, rule *LimitRule, err error) {
	ok, rule, err := i.allow(ctx, target)
	if err != nil {
		return "", nil, err
	}
	if !ok {
		return "", rule, nil
	}
	if err = fn(); err != nil {
		return "", nil, err
	}
	token = idx.UUIDv4()
	key := i.getCaptchaKey(token)
	err = i.rdb.Set(ctx, key, val, time.Duration(i.config.Keeping)*time.Second).Err()
	if err != nil {
		return "", nil, err
	}
	return token, nil, nil
}

func (i *Impl) Verify(ctx context.Context, token, Captcha string) (ok bool, err error) {
	ok, err = i.allowVerify(ctx, token)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, ecode.ErrCaptchaTriesTooMany
	}
	key := i.getCaptchaKey(token)
	storedToken, err := i.rdb.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if i.config.CaseSensitive {
		ok = storedToken == Captcha
	} else {
		ok = strings.ToLower(storedToken) == strings.ToLower(Captcha)
	}
	return ok, nil
}

func (i *Impl) getCaptchaKey(token string) string {
	return fmt.Sprintf("%s:captcha:%s", i.config.KeyPrefix, token)
}

func (i *Impl) allowVerify(ctx context.Context, token string) (ok bool, err error) {
	key := fmt.Sprintf("%s:err:%s", i.config.KeyPrefix, token)
	return i.rdb.AllowFixedLimit(ctx, key, time.Duration(i.config.ErrTryLimit)*time.Second, i.config.Keeping)
}

func (i *Impl) allow(ctx context.Context, target string) (ok bool, rule *LimitRule, err error) {
	ok, w, err := i.limiter.Allow(ctx, target)
	if err != nil {
		return false, nil, err
	}
	if !ok {
		rule = &LimitRule{
			Duration: uint32(w.Duration.Seconds()),
			Limit:    w.Limit,
			Tag:      w.Tag,
		}
		return false, rule, nil
	}
	return true, nil, nil
}
