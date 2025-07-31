package user

import "github.com/byteflowing/base/dal/query"

type Repo interface{}

type GenRepo struct {
	db    *query.Query
	cache Cache
}

func NewStore(db *query.Query, cache Cache) *GenRepo {
	return &GenRepo{
		db:    db,
		cache: cache,
	}
}
