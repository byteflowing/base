package base

import (
	"github.com/byteflowing/base/biz/config"
	"github.com/byteflowing/base/biz/pkg/message"
	"github.com/byteflowing/base/biz/pkg/user"
	"github.com/cloudwego/kitex/pkg/klog"
)

type ServiceOpts struct {
	Config  *config.Config
	Message message.Message
	User    user.User
}

type Service struct {
	config  *config.Config
	Message message.Message
	User    user.User
}

func NewService(opts *ServiceOpts) *Service {
	return &Service{
		config:  opts.Config,
		Message: opts.Message,
		User:    opts.User,
	}
}

func (s *Service) SetLogger(logger klog.FullLogger) {
	klog.SetLogger(logger)
}

func (s *Service) GetConfig() *config.Config {
	return s.config
}
