package user

import (
	"sync"

	"github.com/byteflowing/base/app/user/service"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
)

var (
	userOnce    sync.Once
	userService *service.UserService
)

func NewOnce(cfg *configv1.Config) *service.UserService {
	userOnce.Do(func() {
		userService = service.NewUserService(cfg)
	})
	return userService
}
