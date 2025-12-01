package service

import (
	"github.com/byteflowing/base/app/maps/dal/query"
	"github.com/byteflowing/base/app/maps/migrate"
	"github.com/byteflowing/base/pkg/utils/slicex"
	"github.com/byteflowing/base/singleton"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
)

func NewMapService(cfg *configv1.Config) *MapService {
	rdb := singleton.NewRDB(cfg.Redis)
	db := singleton.NewDB(cfg.Db)
	q := query.Use(db)
	manager := NewMapManager(cfg.Maps.KeyPrefix, rdb, q)
	m := &MapService{
		services: newMapClient(manager, cfg.Maps),
		manager:  manager,
		db:       q,
	}
	if cfg.Maps.AutoMigrate {
		mi := migrate.NewMigrate(db)
		if err := mi.MigrateDB(); err != nil {
			panic(err)
		}
	}
	return m
}

func newMapClient(
	manager *MapManager,
	mapConfig *mapsv1.MapConfig) map[enumv1.MapSource]IMapService {
	enables := slicex.Unique(mapConfig.Enables)
	maps := make(map[enumv1.MapSource]IMapService, len(enables))
	for _, source := range enables {
		switch source {
		case enumv1.MapSource_MAP_SOURCE_AMAP:
			maps[source] = NewAmapImpl(manager)
		case enumv1.MapSource_MAP_SOURCE_TENCENT:
			maps[source] = NewTencent(manager)
		case enumv1.MapSource_MAP_SOURCE_TIAN_DI_TU:
			maps[source] = NewTiandituImpl(manager)
		case enumv1.MapSource_MAP_SOURCE_HUAWEI:
			maps[source] = NewHuaweiImpl(manager)
		}
	}
	return maps
}
