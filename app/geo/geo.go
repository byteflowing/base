package geo

import (
	"sync"

	"github.com/byteflowing/base/app/geo/service"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
)

var (
	geoOnce    sync.Once
	geoService *service.GeoService
)

func NewOnce(cfg *configv1.Config) *service.GeoService {
	geoOnce.Do(func() {
		geoService = NewGeoServiceWithConfig(cfg)
	})
	return geoService
}
