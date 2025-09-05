package common

import (
	configv1 "github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/go-common/redis"
)

func NewRDB(config *configv1.Redis) *redis.Redis {
	return redis.New(&redis.Config{
		Addr:       config.Addr,
		User:       config.User,
		Password:   config.Password,
		DB:         int(config.Db),
		Protocol:   int(config.Protocol),
		ClientName: config.ClientName,
	})
}
