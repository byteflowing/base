package common

import (
	"gorm.io/gorm"

	"github.com/byteflowing/go-common/orm"
	dbv1 "github.com/byteflowing/proto/gen/go/db/v1"
)

func NewDb(config *dbv1.DbConfig) *gorm.DB {
	return orm.New(config)
}
