package maps

import (
	"sync"

	"github.com/byteflowing/base/app/maps/service"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
)

var (
	mapOnce    sync.Once
	mapService *service.MapService
)

func NewOnce(cfg *configv1.Config) *service.MapService {
	mapOnce.Do(func() {
		mapService = service.NewMapService(cfg)
	})
	return mapService
}
