package asynqx

import "github.com/hibiken/asynq"

type Task struct {
	*asynq.Task
}

// NewTask 创建task
// typename与server.RegisterHandler及server.RegisterHandlerFunc中的pattern是关系关系
// 1. 使用*作为匹配符
// 2. 匹配是最长匹配优先
// e.g. 这里填写email:send_welcome RegisterHandler中填写 "email:*" 或者 "email:send_welcome" 则会匹配上
// 如果这里填写email:send_welcome RegisterHandler中有 "email:send_welcome" 和 "email:*" 则会匹配"email:send_welcome"
func NewTask(typename string, payload []byte, opts ...asynq.Option) *Task {
	return &Task{
		asynq.NewTask(typename, payload, opts...),
	}
}
