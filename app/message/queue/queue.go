package queue

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/byteflowing/base/pkg/logx"
	"github.com/byteflowing/base/pkg/queue/asynqx"
	"github.com/byteflowing/base/pkg/redis"
	"github.com/byteflowing/base/singleton"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	"github.com/hibiken/asynq"
)

type Queue struct {
	server *asynqx.Server
	client *asynqx.Client
}

func NewQueue(rdb *redis.Redis, config *configv1.AsynqServerConfig) *Queue {
	return &Queue{
		server: singleton.NewAsynqServer(config, rdb),
		client: singleton.NewAsynqClient(rdb),
	}
}

// EnQueue : 将要发送的消息放入队列
// taskName使用 const中定义的常量字符串格式拼接 e.g. fmt.Sprintf(TaskSendSmsMessage, enumv1.SenderTypeSms, enumv1.SmsVendorAli)
func (q *Queue) EnQueue(ctx context.Context, taskName string, message proto.Message, options ...asynq.Option) error {
	payload, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	task := asynqx.NewTask(taskName, payload, options...)
	_, err = q.client.EnqueueContext(ctx, task)
	logx.Debug(
		"[message queue] EnQueue]",
		zap.String("task", taskName),
		zap.Any("payload", payload),
		zap.Error(err),
	)
	return err
}

func (q *Queue) RegisterHandler(taskName string, handlerFunc asynq.HandlerFunc) {
	q.server.RegisterHandlerFunc(taskName, handlerFunc)
}
