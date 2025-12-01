package asynqx

import (
	"github.com/byteflowing/base/pkg/redis"
	"github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type Server struct {
	server   *asynq.Server
	serveMux *asynq.ServeMux
}

func NewServer(rdbCfg *RedisConfig, config *asynq.Config) *Server {
	return &Server{
		server:   asynq.NewServer(rdbCfg.getOpts(), *config),
		serveMux: asynq.NewServeMux(),
	}
}

func NewServerFromRDB(rdb *redis.Redis, config *asynq.Config) *Server {
	return &Server{
		server:   asynq.NewServerFromRedisClient(rdb.GetUniversalClient(), *config),
		serveMux: asynq.NewServeMux(),
	}
}

// RegisterHandler 注册处理task的对应方法
// 也可以注册实现了ProcessTask的方法
// 这里的pattern与NewTask中的typename有关联关系
// 1. 使用*作为匹配符
// 2. 匹配是最长匹配优先
// e.g. 这里填写"email:*" 或者 "email:send_welcome"  NewTask中email:send_welcome则会匹配上
func (s *Server) RegisterHandler(pattern string, handler asynq.Handler) {
	s.serveMux.Handle(pattern, handler)
}

// RegisterHandlerFunc 同RegisterHandler，传入的函数会自动实现ProcessTask方法
func (s *Server) RegisterHandlerFunc(pattern string, handler asynq.HandlerFunc) {
	s.serveMux.HandleFunc(pattern, handler)
}

// FindHandler 查找task对应的handler
func (s *Server) FindHandler(task *Task) (h asynq.Handler, pattern string) {
	return s.serveMux.Handler(task.Task)
}

// RegisterMiddleware 注册中间件
// 中间件接收handler作为参数，返回一个新的handler
func (s *Server) RegisterMiddleware(middlewares ...asynq.MiddlewareFunc) {
	s.serveMux.Use(middlewares...)
}

func (s *Server) Ping() (err error) {
	return s.server.Ping()
}

// Run 运行并且会阻塞直到收到os.Signal
func (s *Server) Run() error {
	return s.server.Run(s.serveMux)
}

// Start 运行不阻塞
func (s *Server) Start() error {
	return s.server.Start(s.serveMux)
}

// Stop 停止
func (s *Server) Stop() {
	s.server.Stop()
}
