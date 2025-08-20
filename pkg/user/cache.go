package user

import (
	"github.com/byteflowing/go-common/redis"
)

type Cache interface{}

type RedisCache struct {
	rdb *redis.Redis
}

func NewRedisCache(rdb *redis.Redis) *RedisCache {
	return &RedisCache{}
}
