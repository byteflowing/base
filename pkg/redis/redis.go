package redis

import (
	"context"
	"strings"

	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	"github.com/redis/go-redis/v9"
)

var (
	Nil = redis.Nil
)

type Redis struct {
	redis.Cmdable
	client  *redis.Client
	cluster *redis.ClusterClient
}

func New(c *configv1.RedisConfig) *Redis {
	r := &Redis{}
	if c.Type == enumv1.RedisType_REDIS_TYPE_NODE {
		opts := &redis.Options{
			Addr:       c.Hosts[0],
			Username:   c.Username,
			Password:   c.Password,
			DB:         int(c.Db),
			ClientName: c.ClientName,
		}
		client := redis.NewClient(opts)
		r.client = client
		r.Cmdable = client
	} else if c.Type == enumv1.RedisType_REDIS_TYPE_CLUSTER {
		opts := &redis.ClusterOptions{
			Addrs:      c.Hosts,
			Username:   c.Username,
			Password:   c.Password,
			ClientName: c.ClientName,
		}
		cluster := redis.NewClusterClient(opts)
		r.cluster = cluster
		r.Cmdable = cluster
	}
	if err := r.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return r
}

func (r *Redis) EvalShaWithReload(ctx context.Context, sha string, script string, keys []string, args ...interface{}) (any, error) {
	res, err := r.EvalSha(ctx, sha, keys, args...).Result()
	if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT") {
		sha, err = r.ScriptLoad(ctx, script).Result()
		if err != nil {
			return nil, err
		}
		return r.EvalSha(ctx, sha, keys, args...).Result()
	}
	return res, err
}

// GetClient 获取redisV9的client
func (r *Redis) GetClient() *redis.Client {
	return r.client
}

// GetCluster 获取redisV9的集群客户端
func (r *Redis) GetCluster() *redis.ClusterClient {
	return r.cluster
}

// GetUniversalClient 获取通用客户端
func (r *Redis) GetUniversalClient() redis.UniversalClient {
	if r.client != nil {
		return r.client
	}
	return r.cluster
}
