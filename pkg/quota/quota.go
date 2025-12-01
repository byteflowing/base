package quota

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/byteflowing/base/pkg/redis"
	redisv9 "github.com/redis/go-redis/v9"
)

//go:embed script_quota.lua
var scriptQuota string

//go:embed script_decr.lua
var scriptDecr string

//go:embed script_sliding.lua
var scriptSliding string

type Result struct {
	Key        string        // redis中的key
	Allowed    bool          // 是否允许
	Current    int64         // 当前计数
	RetryAfter time.Duration // 多久可以重试
}

type SlidingDetail struct {
	Quota     int
	Used      int
	Window    time.Duration
	RemainTTL time.Duration
}

type Detail struct {
	Used   int
	Window time.Duration
}

func getDetail(ctx context.Context, key string, rdb *redis.Redis) (*Detail, error) {
	var valCmd *redisv9.StringCmd
	var ttlCmd *redisv9.DurationCmd
	_, err := rdb.Pipelined(ctx, func(pipeliner redisv9.Pipeliner) error {
		valCmd = pipeliner.Get(ctx, key)
		ttlCmd = pipeliner.PTTL(ctx, key)
		return nil
	})
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	var used int64
	if valCmd != nil {
		used, err = valCmd.Int64()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
	}
	var ttl time.Duration
	if valCmd != nil {
		ttl, err = ttlCmd.Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
	}
	detail := &Detail{
		Used:   int(used),
		Window: ttl,
	}
	return detail, nil
}

func parseResult(luaResult interface{}, key string) (*Result, error) {
	arr, ok := luaResult.([]interface{})
	if !ok || len(arr) < 3 {
		return nil, errors.New("invalid redis lua response")
	}
	allowed := arr[0].(int64) == 1
	current := arr[1].(int64)
	retryAfterMilli := arr[2].(int64)
	var retryAfter time.Duration
	if retryAfterMilli > 0 {
		retryAfter = time.Duration(retryAfterMilli) * time.Millisecond
	}
	result := &Result{
		Key:        key,
		Allowed:    allowed,
		Current:    current,
		RetryAfter: retryAfter,
	}
	return result, nil
}
