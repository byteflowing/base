package asynqx

import (
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	"github.com/hibiken/asynq"
)

type RedisConfig struct {
	*configv1.RedisConfig
}

func (r *RedisConfig) getOpts() asynq.RedisConnOpt {
	if r.Type == enumv1.RedisType_REDIS_TYPE_NODE {
		return asynq.RedisClientOpt{
			Addr:     r.Hosts[0],
			Username: r.Username,
			Password: r.Password,
			DB:       int(r.Db),
		}
	}
	return asynq.RedisClusterClientOpt{
		Addrs:    r.Hosts,
		Username: r.Username,
		Password: r.Password,
	}
}
