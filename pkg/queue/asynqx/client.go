package asynqx

import (
	"context"

	"github.com/byteflowing/base/pkg/redis"
	"github.com/hibiken/asynq"
)

type Client struct {
	cli *asynq.Client
}

func NewClient(rdbCfg *RedisConfig) *Client {
	cli := asynq.NewClient(rdbCfg.getOpts())
	return &Client{cli}
}

func NewClientFromRDB(rdb *redis.Redis) *Client {
	cli := asynq.NewClientFromRedisClient(rdb.GetUniversalClient())
	return &Client{cli}
}

func (c *Client) Enqueue(task *Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.cli.Enqueue(task.Task, opts...)
}

func (c *Client) EnqueueContext(ctx context.Context, task *Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.cli.EnqueueContext(ctx, task.Task, opts...)
}

func (c *Client) Ping() error {
	return c.cli.Ping()
}
