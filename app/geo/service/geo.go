package service

import (
	"github.com/byteflowing/base/app/geo/migrate"
	"gorm.io/gorm"

	"github.com/byteflowing/base/app/geo/cache"
	"github.com/byteflowing/base/app/geo/dal/query"
	localCache "github.com/byteflowing/base/pkg/cache"
	"github.com/byteflowing/base/pkg/redis"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
)

const (
	resultOK = "OK"
)

type GeoService struct {
	cfg   *geov1.Config
	db    *query.Query
	cache cache.Cache
	geov1.UnimplementedGeoServiceServer
}

func NewGeoService(
	cfg *geov1.Config,
	db *gorm.DB,
	rdb *redis.Redis,
	localCache *localCache.Cache,
) *GeoService {
	_db := query.Use(db)
	_cache := cache.New(cfg.Cache, rdb, localCache)
	s := &GeoService{
		cfg:   cfg,
		db:    _db,
		cache: _cache,
	}
	if cfg.AutoMigrate {
		m := migrate.NewMigrate(db, cfg)
		if err := m.MigrateDB(); err != nil {
			panic(err)
		}
		if err := m.MigrateCountries(cfg.GlobalCountriesCodePath); err != nil {
			panic(err)
		}
		if err := m.MigratePhoneCode(cfg.GlobalPhoneCodePath); err != nil {
			panic(err)
		}
	}
	return s
}
