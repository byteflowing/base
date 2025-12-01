package quota

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/byteflowing/base/pkg/redis"
	"github.com/byteflowing/base/pkg/utils/timex"
)

const (
	calendarKeyFormat = "%s:%s:{%s}"
)

// CalendarPeriod 表示自然周期类型
type CalendarPeriod string

const (
	PeriodHour  CalendarPeriod = "hour"
	PeriodDay   CalendarPeriod = "day"
	PeriodMonth CalendarPeriod = "month"
	PeriodYear  CalendarPeriod = "year"
)

type CalendarQuota struct {
	rdb *redis.Redis
}

func NewCalendarQuota(rdb *redis.Redis) *CalendarQuota {
	return &CalendarQuota{rdb: rdb}
}

func (c *CalendarQuota) Allow(ctx context.Context, keyPrefix, target string, quota int, period CalendarPeriod) (*Result, error) {
	return c.allow(ctx, keyPrefix, target, quota, period, 1)
}

func (c *CalendarQuota) AllowN(ctx context.Context, keyPrefix, target string, quota int, period CalendarPeriod, n int) (*Result, error) {
	return c.allow(ctx, keyPrefix, target, quota, period, n)
}

// Decr : 在使用调用本方法时，前一窗口可能已经过期，减去的实际上是当前窗口的数量，allow返回的Resu中有expiresAt字段，可以用来判断当前窗口是否过期
func (c *CalendarQuota) Decr(ctx context.Context, keyPrefix, target string, period CalendarPeriod) (newValue int64, err error) {
	return c.decr(ctx, keyPrefix, target, period, 1)
}

// DecrN : 在使用调用本方法时，前一窗口可能已经过期，减去的实际上是当前窗口的数量，allow返回的Resu中有expiresAt字段，可以用来判断当前窗口是否过期
func (c *CalendarQuota) DecrN(ctx context.Context, keyPrefix, target string, period CalendarPeriod, n int) (newValue int64, err error) {
	return c.decr(ctx, keyPrefix, target, period, n)
}

func (c *CalendarQuota) Reset(ctx context.Context, keyPrefix, target string, period CalendarPeriod) (err error) {
	key := c.getKey(keyPrefix, target, period)
	return c.rdb.Del(ctx, key).Err()
}

func (c *CalendarQuota) GetDetail(ctx context.Context, keyPrefix, target string, period CalendarPeriod) (*Detail, error) {
	key := c.getKey(keyPrefix, target, period)
	return getDetail(ctx, key, c.rdb)
}

func (c *CalendarQuota) allow(ctx context.Context, keyPrefix, target string, quota int, period CalendarPeriod, n int) (*Result, error) {
	ttl := c.remainingMillis(period)
	randTTL := c.jitterMillis(period)
	key := c.getKey(keyPrefix, target, period)
	res, err := c.rdb.Eval(ctx, scriptQuota, []string{key}, ttl+randTTL, quota, n).Result()
	if err != nil {
		return nil, err
	}
	result, err := parseResult(res, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *CalendarQuota) decr(ctx context.Context, keyPrefix, target string, period CalendarPeriod, n int) (int64, error) {
	key := c.getKey(keyPrefix, target, period)
	return c.rdb.Eval(ctx, scriptDecr, []string{key}, n).Int64()
}

func (c *CalendarQuota) getKey(keyPrefix, target string, period CalendarPeriod) string {
	periodID := c.periodID(period)
	return fmt.Sprintf(calendarKeyFormat, keyPrefix, periodID, target)
}

func (c *CalendarQuota) remainingMillis(period CalendarPeriod) int64 {
	switch period {
	case PeriodHour:
		return timex.EndOfHourMillis()
	case PeriodDay:
		return timex.EndOfDayMillis()
	case PeriodMonth:
		return timex.EndOfMonthMillis()
	case PeriodYear:
		return timex.EndOfYearMillis()
	default:
		panic("unsupported period")
	}
}

func (c *CalendarQuota) periodID(period CalendarPeriod) string {
	now := time.Now()
	switch period {
	case PeriodHour:
		return now.Format("2006010215")
	case PeriodDay:
		return now.Format("20060102")
	case PeriodMonth:
		return now.Format("200601")
	case PeriodYear:
		return now.Format("2006")
	default:
		panic("unsupported period type")
	}
}

func (c *CalendarQuota) jitterMillis(period CalendarPeriod) int64 {
	var maxDuration time.Duration
	switch period {
	case PeriodHour:
		maxDuration = time.Minute
	case PeriodDay:
		maxDuration = 2 * time.Minute
	case PeriodMonth:
		maxDuration = 10 * time.Minute
	case PeriodYear:
		maxDuration = time.Hour
	default:
		maxDuration = 2 * time.Minute
	}
	return fastrand.Int63n(int64(maxDuration.Milliseconds()) + 1)
}
