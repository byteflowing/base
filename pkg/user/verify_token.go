package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/byteflowing/base/ecode"
	configv1 "github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/go-common/redis"
)

type TokenVerifier interface {
	Store(ctx context.Context, token string, uid int64) error
	GetUid(ctx context.Context, token string) (uid int64, err error)
}

type RedisTokenVerifier struct {
	config *configv1.TokenVerify
	rdb    *redis.Redis
}

func NewRedisTokenVerifier(config *configv1.TokenVerify) TokenVerifier {
	return &RedisTokenVerifier{
		config: config,
	}
}

func (r *RedisTokenVerifier) Store(ctx context.Context, token string, uid int64) error {
	return r.rdb.Set(ctx, r.key(token), uid, r.config.Keeping.AsDuration()).Err()
}

func (r *RedisTokenVerifier) GetUid(ctx context.Context, token string) (uid int64, err error) {
	uid, err = r.rdb.Get(ctx, r.key(token)).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ecode.ErrCaptchaNotExist
		}
		return 0, err
	}
	return
}

func (r *RedisTokenVerifier) key(token string) string {
	return fmt.Sprintf("%s:%s", r.config.Prefix, token)
}
