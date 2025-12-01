package singleton

import (
	"log"
	"sync"

	"gorm.io/gorm"

	"github.com/byteflowing/base/pkg/cache"
	"github.com/byteflowing/base/pkg/config"
	"github.com/byteflowing/base/pkg/cron"
	"github.com/byteflowing/base/pkg/db"
	"github.com/byteflowing/base/pkg/queue/asynqx"
	"github.com/byteflowing/base/pkg/redis"
	"github.com/byteflowing/base/pkg/shortid"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	"github.com/hibiken/asynq"
)

var (
	dbOnce             sync.Once
	rdbOnce            sync.Once
	cronOnce           sync.Once
	configOnce         sync.Once
	localCacheOnce     sync.Once
	shortIDOnce        sync.Once
	asynqServerOnce    sync.Once
	asynqClientOnce    sync.Once
	asynqSchedulerOnce sync.Once
)

var (
	rdb            *redis.Redis
	_db            *gorm.DB
	_cron          *cron.Cron
	_localCache    *cache.Cache
	_config        *configv1.Config
	_shortID       *shortid.Generator
	asynqServer    *asynqx.Server
	asynqClient    *asynqx.Client
	asynqScheduler *asynqx.Scheduler
)

func NewConfig(file string) *configv1.Config {
	configOnce.Do(func() {
		log.Println("Loading config file:", file)
		_config = &configv1.Config{}
		if err := config.ReadProtoConfig(file, _config); err != nil {
			panic(err)
		}
	})
	return _config
}

func NewDB(config *configv1.DbConfig) *gorm.DB {
	dbOnce.Do(func() {
		_db = db.New(config)
	})
	return _db
}

func NewRDB(config *configv1.RedisConfig) *redis.Redis {
	rdbOnce.Do(func() {
		rdb = redis.New(config)
	})
	return rdb
}

func NewShortID(config *configv1.ShortId) *shortid.Generator {
	shortIDOnce.Do(func() {
		var err error
		_shortID, err = shortid.NewShortIdGenerator(&shortid.Config{
			Alphabet:  config.Alphabet,
			MinLength: uint8(config.MinLength),
			Blocklist: config.BlockList,
		})
		if err != nil {
			panic(err)
		}
	})
	return _shortID
}

func NewAsynqServer(config *configv1.AsynqServerConfig, rdb *redis.Redis) *asynqx.Server {
	asynqServerOnce.Do(func() {
		c := &asynq.Config{
			Concurrency:       int(config.Concurrency),
			TaskCheckInterval: config.TaskCheckInterval.AsDuration(),
			Queues: map[string]int{
				"critical": int(config.QueueCriticalPriority),
				"default":  int(config.QueueDefaultPriority),
				"low":      int(config.QueueLowPriority),
			},
			ShutdownTimeout:          config.ShutdownTimeout.AsDuration(),
			HealthCheckInterval:      config.HealthCheckInterval.AsDuration(),
			DelayedTaskCheckInterval: config.DelayTaskCheckInterval.AsDuration(),
			GroupGracePeriod:         config.GroupGracePeriod.AsDuration(),
			GroupMaxDelay:            config.GroupMaxDelay.AsDuration(),
			GroupMaxSize:             int(config.GroupMaxSize),
			JanitorInterval:          config.JanitorInterval.AsDuration(),
			JanitorBatchSize:         int(config.JanitorBatchSize),
		}
		asynqServer = asynqx.NewServerFromRDB(rdb, c)
		addStarter(asynqServer)
	})
	return asynqServer
}

func NewAsynqClient(rdb *redis.Redis) *asynqx.Client {
	asynqClientOnce.Do(func() {
		asynqx.NewClientFromRDB(rdb)
	})
	return asynqClient
}

func NewAsynqScheduler(config *configv1.AsynqSchedulerConfig, rdb *redis.Redis) *asynqx.Scheduler {
	asynqSchedulerOnce.Do(func() {
		asynqScheduler = asynqx.NewSchedulerFromRDB(
			rdb,
			&asynq.SchedulerOpts{
				HeartbeatInterval: config.HealthCheckInterval.AsDuration(),
			},
		)
		addStarter(asynqScheduler)
	})
	return asynqScheduler
}

func NewCron() *cron.Cron {
	cronOnce.Do(func() {
		_cron = cron.New()
		addStarter(_cron)
	})
	return _cron
}

func NewLocalCache(config *configv1.LocalCacheConfig) *cache.Cache {
	if config == nil {
		return nil
	}
	localCacheOnce.Do(func() {
		_localCache = cache.New(config)
	})
	return _localCache
}
