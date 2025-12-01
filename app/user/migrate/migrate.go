package migrate

import (
	"gorm.io/gorm"

	"github.com/byteflowing/base/app/user/dal/model"
)

type Migrate struct {
	_db *gorm.DB
}

func NewMigrate(db *gorm.DB) *Migrate {
	return &Migrate{
		_db: db,
	}
}

func (m *Migrate) MigrateDB() (err error) {
	return m._db.AutoMigrate(
		&model.UserAccount{},
		&model.UserAuth{},
		&model.UserSignLog{},
	)
}
