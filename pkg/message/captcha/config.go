package captcha

import (
	"time"

	"github.com/byteflowing/go-common/redis"
)

type Config struct {
	KeyPrefix     string       // 验证码在redis中存储的前缀
	Keeping       uint32       // 验证码保留时长，单位s
	CaseSensitive bool         // 大小写敏感
	ErrTryLimit   uint32       // 可以尝试几次验证，超过后删除验证码
	Limits        []*LimitRule // 验证码短信限流规则
}

func (c *Config) ToWindows() []*redis.Window {
	var windows []*redis.Window
	for _, l := range c.Limits {
		windows = append(windows, &redis.Window{
			Duration: time.Duration(l.Duration) * time.Second,
			Limit:    l.Limit,
			Tag:      l.Tag,
		})
	}
	return windows
}

type LimitRule struct {
	Duration   uint32 // 短信发送限制时间周期，单位s
	Limit      uint32 // 周期内的限制次数
	Tag        string // 限制的标签，方便识别
	RetryAfter uint64 // 作为返回值有效，还有多少秒可以重试
}
