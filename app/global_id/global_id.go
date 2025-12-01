package global_id

import (
	"sync"

	"github.com/byteflowing/base/app/global_id/service"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
)

var (
	globalIDOnce    sync.Once
	globalIDService *service.GlobalIDService
)

func NewOnce(cfg *configv1.Config) *service.GlobalIDService {
	globalIDOnce.Do(func() {
		globalIDService = service.NewGlobalIDService(cfg)
	})
	return globalIDService
}
