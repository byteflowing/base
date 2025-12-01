package timex

import (
	"time"
)

const (
	YYYYMMDDFormat = "20060102"
)

func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfHourMillis() int64 {
	now := time.Now()
	end := now.Truncate(time.Hour).Add(time.Hour)
	return end.Sub(now).Milliseconds()
}

func EndOfDayMillis() int64 {
	now := time.Now()
	// 截断到当天起点（00:00:00），加1天即为当天结束点
	end := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	return end.Sub(now).Milliseconds()
}

func EndOfMonthMillis() int64 {
	now := time.Now()
	// 当月第一天 +1个月 → 次月第一天（即当月结束点）
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := firstDayOfMonth.AddDate(0, 1, 0) // 加1个月
	return end.Sub(now).Milliseconds()
}

func EndOfYearMillis() int64 {
	now := time.Now()
	// 当年第一天 +1年 → 次年第一天（即当年结束点）
	firstDayOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	end := firstDayOfYear.AddDate(1, 0, 0) // 加1年
	return end.Sub(now).Milliseconds()
}

// GetDayString 获取当天日期字符串格式：20060102
// offset为加减多少天，0为当天
func GetDayString(offset int) string {
	return time.Now().AddDate(0, 0, offset).Format(YYYYMMDDFormat)
}
