package asynqx

import (
	"github.com/byteflowing/base/pkg/redis"
	"github.com/hibiken/asynq"
)

type Scheduler struct {
	s *asynq.Scheduler
}

func NewScheduler(rdbCfg *RedisConfig, opts *asynq.SchedulerOpts) *Scheduler {
	s := asynq.NewScheduler(rdbCfg.getOpts(), opts)
	return &Scheduler{s}
}

func NewSchedulerFromRDB(rdb *redis.Redis, opts *asynq.SchedulerOpts) *Scheduler {
	s := asynq.NewSchedulerFromRedisClient(rdb.GetUniversalClient(), opts)
	return &Scheduler{s}
}

func (s *Scheduler) Register(cronSpec string, task *Task, opts ...asynq.Option) (entryID string, err error) {
	return s.s.Register(cronSpec, task.Task, opts...)
}

func (s *Scheduler) Unregister(entryID string) error {
	return s.s.Unregister(entryID)
}

func (s *Scheduler) Run() error {
	return s.s.Run()
}

func (s *Scheduler) Start() error {
	return s.s.Start()
}

func (s *Scheduler) Stop() {
	s.s.Shutdown()
}

func (s *Scheduler) Ping() error {
	return s.s.Ping()
}
