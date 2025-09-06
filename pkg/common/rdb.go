package common

import (
	"github.com/byteflowing/go-common/redis"
	dbv1 "github.com/byteflowing/proto/gen/go/db/v1"
)

func NewRDB(config *dbv1.RedisConfig) *redis.Redis {
	return redis.New(config)
}
