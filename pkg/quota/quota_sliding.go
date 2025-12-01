package quota

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/byteflowing/base/pkg/redis"
	redisv9 "github.com/redis/go-redis/v9"
)

const (
	slidingKeyFormat = "%s:{%s}:%d"
)

type SlidingRule struct {
	Quota  int
	Window time.Duration
}

type SlidingRuleResult struct {
	Allowed    bool
	Current    int32
	RetryAfter time.Duration
	Rule       *SlidingRule
}

type SlidingQuota struct {
	rdb *redis.Redis
}

func NewSlidingQuota(rdb *redis.Redis) *SlidingQuota {
	return &SlidingQuota{rdb: rdb}
}

func (s *SlidingQuota) Allow(ctx context.Context, keyPrefix, target string, rules []*SlidingRule) (*SlidingRuleResult, error) {
	var keys = make([]string, 0, len(rules))
	var args = make([]interface{}, 0, len(rules)*2)
	for _, rule := range rules {
		keys = append(keys, s.getKey(keyPrefix, target, rule.Window))
		args = append(args, rule.Window.Milliseconds(), rule.Quota)
	}
	res, err := s.rdb.Eval(ctx, scriptSliding, keys, args...).Result()
	if err != nil {
		return nil, err
	}
	return s.parseResult(res, rules)
}

func (s *SlidingQuota) Reset(ctx context.Context, keyPrefix, target string, rules []*SlidingRule) error {
	keys := make([]string, 0, len(rules))
	for _, rule := range rules {
		keys = append(keys, s.getKey(keyPrefix, target, rule.Window))
	}
	_, err := s.rdb.Del(ctx, keys...).Result()
	return err
}

func (s *SlidingQuota) GetDetail(ctx context.Context, keyPrefix, target string, rules []*SlidingRule) ([]*SlidingDetail, error) {
	pipe := s.rdb.Pipeline()
	counts := len(rules)
	stringCmds := make([]*redisv9.StringCmd, 0, counts)
	durationCmds := make([]*redisv9.DurationCmd, 0, counts)
	keys := make([]string, 0, counts)
	for _, rule := range rules {
		key := s.getKey(keyPrefix, target, rule.Window)
		keys = append(keys, key)
		stringCmds = append(stringCmds, pipe.Get(ctx, key))
		durationCmds = append(durationCmds, pipe.PTTL(ctx, key))
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	details := make([]*SlidingDetail, 0, counts)
	for idx, rule := range rules {
		used, err := stringCmds[idx].Int64()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
		remainTTL, err := durationCmds[idx].Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
		details = append(details, &SlidingDetail{
			Quota:     rule.Quota,
			Used:      int(used),
			Window:    rule.Window,
			RemainTTL: remainTTL,
		})
	}
	return details, nil
}

func (s *SlidingQuota) getKey(keyPrefix, target string, window time.Duration) string {
	return fmt.Sprintf(slidingKeyFormat, keyPrefix, target, window.Milliseconds())
}

func (s *SlidingQuota) parseResult(luaResult interface{}, rules []*SlidingRule) (*SlidingRuleResult, error) {
	arr, ok := luaResult.([]interface{})
	if !ok || len(arr) < 4 {
		return nil, errors.New("invalid redis lua response")
	}
	allowed := arr[0].(int64) == 1
	retryAfterMilli := arr[1].(int64)
	ruleIdx := arr[2].(int64)
	current := arr[3].(int64)
	var retryAfter time.Duration
	var rule *SlidingRule
	if retryAfterMilli > 0 {
		retryAfter = time.Duration(retryAfterMilli) * time.Millisecond
		rule = rules[ruleIdx]
	}
	result := &SlidingRuleResult{
		Allowed:    allowed,
		Current:    int32(current),
		RetryAfter: retryAfter,
		Rule:       rule,
	}
	return result, nil
}
