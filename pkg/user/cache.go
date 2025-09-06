package user

import (
	"github.com/byteflowing/go-common/redis"
	userv1 "github.com/byteflowing/proto/gen/go/services/user/v1"
)

type Cache interface{}

type CacheImpl struct {
	rdb *redis.Redis
}

func NewCache(c *userv1.UserCache, rdb *redis.Redis) Cache {
	return &CacheImpl{}
}
