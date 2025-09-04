package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/byteflowing/base/ecode"
	configv1 "github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/go-common/redis"
)

type TowStepVerifier struct {
	config *configv1.TokenVerify
	rdb    *redis.Redis
}

func NewTwoStepVerifier(config *configv1.TokenVerify, rdb *redis.Redis) *TowStepVerifier {
	return &TowStepVerifier{
		config: config,
		rdb:    rdb,
	}
}

func (r *TowStepVerifier) Store(ctx context.Context, token string, uid int64) error {
	return r.rdb.Set(ctx, r.key(token), uid, r.config.Keeping.AsDuration()).Err()
}

func (r *TowStepVerifier) Verify(ctx context.Context, token string) (uid int64, err error) {
	uid, err = r.rdb.Get(ctx, r.key(token)).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ecode.ErrCaptchaNotExist
		}
		return 0, err
	}
	return
}

func (r *TowStepVerifier) key(token string) string {
	return fmt.Sprintf("%s:%s", r.config.Prefix, token)
}
