package message

import (
	"sync"

	"github.com/byteflowing/base/app/message/service"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
)

var (
	messageOnce    sync.Once
	messageService *service.MessageService
)

func NewOnce(cfg *configv1.Config) *service.MessageService {
	messageOnce.Do(func() {
		messageService = NewMessageServiceWithConfig(cfg)
	})
	return messageService
}
