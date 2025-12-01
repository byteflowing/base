//go:build wireinject
// +build wireinject

package geo

import (
	"github.com/byteflowing/base/app/geo/service"
	"github.com/byteflowing/base/singleton"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	"github.com/google/wire"
)

var providerSet = wire.NewSet(
	singleton.NewDB,
	singleton.NewRDB,
	singleton.NewLocalCache,
	service.NewGeoService,
	wire.FieldsOf(new(*configv1.Config), "Db", "Redis", "Geo", "LocalCache"),
)

func NewGeoServiceWithConfig(cfg *configv1.Config) *service.GeoService {
	panic(wire.Build(providerSet))
}
