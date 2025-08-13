package user

import (
	"github.com/byteflowing/go-common/redis"
)

type Cache interface{}

type RedisCache struct {
	rdb    *redis.Redis
	config *Config
}

func NewRedisCache(rdb *redis.Redis) *RedisCache {
	return &RedisCache{}
}
