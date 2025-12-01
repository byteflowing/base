package migrate

import (
	"github.com/byteflowing/base/app/geo/dal/model"
	geov1 "github.com/byteflowing/proto/gen/go/geo/v1"
	"gorm.io/gorm"

	"github.com/byteflowing/base/app/geo/dal/query"
)

type Migrate struct {
	_db    *gorm.DB
	_query *query.Query
	cfg    *geov1.Config
}

func NewMigrate(db *gorm.DB, cfg *geov1.Config) *Migrate {
	return &Migrate{
		_db:    db,
		_query: query.Use(db),
		cfg:    cfg,
	}
}

func (m *Migrate) MigrateDB() error {
	return m._db.Migrator().AutoMigrate(
		&model.GeoPhoneCode{},
		&model.GeoCountry{},
		&model.GeoRegion{},
	)
}
