package asynqx

import (
	"github.com/byteflowing/base/pkg/redis"
	"github.com/hibiken/asynq"
)

type Inspector struct {
	*asynq.Inspector
}

func NewInspector(rdbCfg *RedisConfig) *Inspector {
	i := asynq.NewInspector(rdbCfg.getOpts())
	return &Inspector{i}
}

func NewInspectorFromRDB(rdb *redis.Redis) *Inspector {
	i := asynq.NewInspectorFromRedisClient(rdb.GetUniversalClient())
	return &Inspector{i}
}
