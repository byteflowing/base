package service

import (
	"github.com/byteflowing/base/biz/pkg/message"
	"github.com/byteflowing/base/biz/pkg/user"
)

type Service interface {
	message.Message
	user.User
}
