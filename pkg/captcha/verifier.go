package captcha

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/byteflowing/base/pkg/redis"
)

//go:embed script_verify_with_fail.lua
var scriptValueVerifier string

const (
	errSuffix = "fails"
)

type ValueVerifyCode int

const (
	ValueVerifySuccess ValueVerifyCode = iota
	ValueVerifyKeyNotFound
	ValueVerifyValueMismatch
	ValueVerifyMaxFails
	ValueVerifyUnknown
)

type ValueVerifyResult struct {
	Code  ValueVerifyCode
	Fails int32
}

type ValueVerifier struct {
	rdb      *redis.Redis
	ttl      time.Duration
	maxTries int
}

// NewValueVerifier 验证值是否匹配并记录错误次数
// 应用场景：验证手机验证码，并在错误时记录错误次数（可以在业务层根据错误次数主动调用Delete删除验证码）
// 验证成功会主动删除验证码
func NewValueVerifier(rdb *redis.Redis, ttl time.Duration, maxTries int) *ValueVerifier {
	return &ValueVerifier{
		rdb:      rdb,
		ttl:      ttl,
		maxTries: maxTries,
	}
}

func (v *ValueVerifier) Save(ctx context.Context, key, value string) error {
	return v.rdb.Set(ctx, key, value, v.ttl).Err()
}

// Verify 主要用于验证存储在redis中的key和value与传递过来的是否一致
// 如果不一致会记录failKey的错误次数，通过lua脚本保证原子操作
func (v *ValueVerifier) Verify(ctx context.Context, key, value string) (result *ValueVerifyResult, err error) {
	failKey := key + ":" + errSuffix
	keys := []string{key, failKey}
	res, err := v.rdb.Eval(ctx, scriptValueVerifier, keys, value, v.maxTries).Result()
	if err != nil {
		return nil, err
	}
	return v.parseResult(res)
}

func (v *ValueVerifier) parseResult(res interface{}) (result *ValueVerifyResult, err error) {
	arr, ok := res.([]interface{})
	if !ok || len(arr) < 2 {
		return nil, fmt.Errorf("unexpected lua result: %#v", res)
	}
	code, _ := arr[0].(int64)
	fails, _ := arr[1].(int64)
	var c ValueVerifyCode
	switch code {
	case 0:
		c = ValueVerifySuccess
	case -1:
		c = ValueVerifyKeyNotFound
	case -2:
		c = ValueVerifyMaxFails
	case -3:
		c = ValueVerifyValueMismatch
	default:
		c = ValueVerifyUnknown
	}
	result = &ValueVerifyResult{
		Code:  c,
		Fails: int32(fails),
	}
	return
}
