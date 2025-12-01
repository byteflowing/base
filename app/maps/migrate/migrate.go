package migrate

import (
	"gorm.io/gorm"

	"github.com/byteflowing/base/app/maps/dal/model"
)

type Migrate struct {
	_db *gorm.DB
}

func NewMigrate(db *gorm.DB) *Migrate {
	return &Migrate{
		_db: db,
	}
}

func (m *Migrate) MigrateDB() error {
	return m._db.Migrator().AutoMigrate(
		&model.MapAccount{},
		&model.MapInterface{},
		&model.MapInterfaceCount{},
	)
}
