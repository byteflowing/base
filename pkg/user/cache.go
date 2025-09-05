package user

import (
	configv1 "github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/go-common/redis"
)

type Cache interface{}

type CacheImpl struct {
	rdb *redis.Redis
}

func NewCache(c *configv1.UserCache, rdb *redis.Redis) Cache {
	return &CacheImpl{}
}
