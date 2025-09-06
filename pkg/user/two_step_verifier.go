package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/go-common/redis"
	userv1 "github.com/byteflowing/proto/gen/go/services/user/v1"
)

type TwoStepVerifier struct {
	config *userv1.TokenVerify
	rdb    *redis.Redis
}

func NewTwoStepVerifier(config *userv1.TokenVerify, rdb *redis.Redis) *TwoStepVerifier {
	return &TwoStepVerifier{
		config: config,
		rdb:    rdb,
	}
}

func (t *TwoStepVerifier) Store(ctx context.Context, token string, uid int64) error {
	return t.rdb.Set(ctx, t.key(token), uid, t.config.Keeping.AsDuration()).Err()
}

func (t *TwoStepVerifier) Verify(ctx context.Context, token string) (uid int64, err error) {
	uid, err = t.rdb.Get(ctx, t.key(token)).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ecode.ErrCaptchaNotExist
		}
		return 0, err
	}
	return
}

func (t *TwoStepVerifier) Delete(ctx context.Context, token string) error {
	return t.rdb.Del(ctx, t.key(token)).Err()
}

func (t *TwoStepVerifier) key(token string) string {
	return fmt.Sprintf("%s:%s", t.config.Prefix, token)
}
