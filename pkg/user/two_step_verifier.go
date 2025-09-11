package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/go-common/redis"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	userv1 "github.com/byteflowing/proto/gen/go/services/user/v1"
)

type TwoStepVerifier struct {
	config *userv1.TokenVerifyConfig
	rdb    *redis.Redis
}

func NewTwoStepVerifier(config *userv1.TokenVerifyConfig, rdb *redis.Redis) *TwoStepVerifier {
	return &TwoStepVerifier{
		config: config,
		rdb:    rdb,
	}
}

func (t *TwoStepVerifier) Store(
	ctx context.Context,
	token string,
	uid int64,
	sender enumsv1.MessageSenderType,
	captchaType string,
) error {
	return t.rdb.Set(ctx, t.key(token, sender, captchaType), uid, t.config.Keeping.AsDuration()).Err()
}

func (t *TwoStepVerifier) Verify(
	ctx context.Context,
	token string,
	sender enumsv1.MessageSenderType,
	captchaType string,
) (uid int64, err error) {
	uid, err = t.rdb.Get(ctx, t.key(token, sender, captchaType)).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ecode.ErrCaptchaNotExist
		}
		return 0, err
	}
	return
}

func (t *TwoStepVerifier) Delete(
	ctx context.Context,
	token string,
	sender enumsv1.MessageSenderType,
	captchaType string,
) error {
	return t.rdb.Del(ctx, t.key(token, sender, captchaType)).Err()
}

func (t *TwoStepVerifier) key(
	token string,
	sender enumsv1.MessageSenderType,
	captchaType string,
) string {
	return fmt.Sprintf("%s:%d:%s:{%s}", t.config.Prefix, sender, captchaType, token)
}
