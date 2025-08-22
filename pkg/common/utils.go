package common

import (
	commonv1 "github.com/byteflowing/base/gen/common/v1"
	"github.com/byteflowing/go-common/redis"
)

func ConvertLimitsToWindows(limits []*commonv1.LimitRule) []*redis.Window {
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
