package db

import (
	"errors"

	"gorm.io/gorm"
)

type PageResult[T any] struct {
	List       []*T
	Page       uint32
	PageSize   uint32
	Total      uint64
	TotalPages uint32
}

// Paginate 带total分页查询的封装
// 需要在传递tx时将其他查询条件先构建到tx中
func Paginate[T any](tx *gorm.DB, page, pageSize uint32) (*PageResult[T], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	var total int64
	if err := tx.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}
	var list []*T
	if total > 0 {
		if err := tx.Session(&gorm.Session{}).
			Offset(int((page - 1) * pageSize)).
			Limit(int(pageSize)).
			Find(&list).Error; err != nil {
			return nil, err
		}
	} else {
		list = []*T{}
	}
	totalPages := uint32((total + int64(pageSize) - 1) / int64(pageSize))
	return &PageResult[T]{
		List:       list,
		Page:       page,
		PageSize:   pageSize,
		Total:      uint64(total),
		TotalPages: totalPages,
	}, nil
}

func IsDBNotFoundErr(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
