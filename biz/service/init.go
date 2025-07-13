package service

import (
	"github.com/byteflowing/base/biz/config"
	"github.com/byteflowing/base/biz/pkg/message"
	"github.com/byteflowing/base/biz/pkg/user"
)

type Opts struct {
	Conf    *config.Config
	Message message.Message
	User    user.User
}

type Impl struct {
	user.User
	message.Message
}

func New(opts *Opts) Service {
	return &Impl{
		opts.User,
		opts.Message,
	}
}
