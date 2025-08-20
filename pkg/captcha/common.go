package captcha

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/byteflowing/base/ecode"
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	configv1 "github.com/byteflowing/base/gen/config/v1"
	enumsV1 "github.com/byteflowing/base/gen/enums/v1"
	"github.com/byteflowing/go-common/idx"
	"github.com/byteflowing/go-common/redis"
	"github.com/byteflowing/go-common/trans"
	"google.golang.org/protobuf/types/known/durationpb"
)

type captcha struct {
	config  *configv1.Captcha
	rdb     *redis.Redis
	limiter *redis.Limiter
}

func newCaptcha(rdb *redis.Redis, c *configv1.Captcha) *captcha {
	limiter := redis.NewLimiter(rdb, c.Prefix, convertLimitsToWindows(c.Limits))
	return &captcha{
		config:  c,
		rdb:     rdb,
		limiter: limiter,
	}
}

func (c *captcha) send(ctx context.Context, target, val string, fn func() error) (token string, rule *commonv1.LimitRule, err error) {
	ok, rule, err := c.allow(ctx, target)
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
	key := c.getCaptchaKey(token)
	v := fmt.Sprintf("%s:%s", target, val)
	err = c.rdb.Set(ctx, key, v, c.config.Keeping.AsDuration()).Err()
	if err != nil {
		return "", nil, err
	}
	return token, nil, nil
}

func (c *captcha) verify(ctx context.Context, target, token, captcha string, sender enumsV1.MessageSenderType) (ok bool, err error) {
	ok, err = c.allowVerify(ctx, token)
	if err != nil {
		return false, err
	}
	key := c.getCaptchaKey(token)
	if !ok {
		// 错误次数过多删除验证码
		if err = c.rdb.Del(ctx, key).Err(); err != nil {
			return false, err
		}
		return false, ecode.ErrCaptchaTooManyErrors
	}
	storedToken, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, ecode.ErrCaptchaNotExist
		}
		return false, err
	}
	vals := strings.Split(storedToken, ":")
	if len(vals) != 2 {
		return false, ecode.ErrInternal
	}
	if vals[0] != target {
		return false, c.getTargetErr(sender)
	}
	if c.config.CaseSensitive {
		ok = storedToken == captcha
	} else {
		ok = strings.ToLower(storedToken) == strings.ToLower(captcha)
	}
	// 验证成功删除验证码
	if ok {
		if err = c.rdb.Del(ctx, key).Err(); err != nil {
			return false, err
		}
	}
	return ok, nil
}

func (c *captcha) getCaptchaKey(token string) string {
	return fmt.Sprintf("%s:{%s}", c.config.Prefix, token)
}

func (c *captcha) getCaptchaErrKey(token string) string {
	return fmt.Sprintf("%s:{%s}", c.config.ErrPrefix, token)
}

func (c *captcha) allowVerify(ctx context.Context, token string) (ok bool, err error) {
	key := c.getCaptchaErrKey(token)
	return c.rdb.AllowFixedLimit(ctx, key, c.config.Keeping.AsDuration(), uint32(c.config.ErrTryLimit))
}

func (c *captcha) allow(ctx context.Context, target string) (ok bool, rule *commonv1.LimitRule, err error) {
	ok, w, after, err := c.limiter.Allow(ctx, target)
	if err != nil {
		return false, nil, err
	}
	if !ok {
		rule = &commonv1.LimitRule{
			Duration:   durationpb.New(w.Duration),
			Limit:      int32(w.Limit),
			Tag:        w.Tag,
			RetryAfter: trans.Ref(after),
		}
		return false, rule, nil
	}
	return true, nil, nil
}

func convertLimitsToWindows(limits []*commonv1.LimitRule) []*redis.Window {
	var windows []*redis.Window
	for _, limit := range limits {
		windows = append(windows, &redis.Window{
			Duration: limit.Duration.AsDuration(),
			Limit:    uint32(limit.Limit),
			Tag:      limit.Tag,
		})
	}
	return windows
}

func (c *captcha) getTargetErr(sender enumsV1.MessageSenderType) error {
	switch sender {
	case enumsV1.MessageSenderType_MESSAGE_SENDER_TYPE_SMS:
		return ecode.ErrPhoneNotMatch
	case enumsV1.MessageSenderType_MESSAGE_SENDER_TYPE_MAIL:
		return ecode.ErrEmailNotMatch
	}
	return ecode.ErrInternal
}
