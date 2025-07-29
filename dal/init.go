package dal

import (
	"github.com/byteflowing/base/dal/query"
	"gorm.io/gorm"
)

func New(orm *gorm.DB) *query.Query {
	return query.Use(orm)
}
