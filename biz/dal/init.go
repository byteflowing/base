package dal

import (
	"github.com/byteflowing/base/biz/dal/query"
	"gorm.io/gorm"
)

func New(orm *gorm.DB) *query.Query {
	return query.Use(orm)
}
