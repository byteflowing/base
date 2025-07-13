//go:build wireinject
// +build wireinject

package base

import (
	"github.com/byteflowing/base/biz/config"
	"github.com/byteflowing/base/biz/dal"
	"github.com/byteflowing/base/biz/pkg/message"
	"github.com/byteflowing/base/biz/pkg/user"
	"github.com/byteflowing/base/biz/service"
	"github.com/byteflowing/go-common/orm"
	"github.com/byteflowing/go-common/redis"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var providerSet = wire.NewSet(
	redis.New,
	orm.New,
	dal.New,
	config.New,
	service.New,
	message.New,
	user.New,
	wire.Struct(new(service.Opts), "*"),
	wire.Struct(new(user.Opts), "*"),
	wire.Struct(new(message.Opts), "*"),
	wire.FieldsOf(new(*config.Config), "DB", "Redis", "Message", "User"),
)

var providerSet2 = wire.NewSet(
	dal.New,
	service.New,
	message.New,
	user.New,
	wire.Struct(new(service.Opts), "*"),
	wire.Struct(new(user.Opts), "*"),
	wire.Struct(new(message.Opts), "*"),
	wire.FieldsOf(new(*config.Config), "User", "Message"),
)

func New(confFile string) service.Service {
	panic(wire.Build(providerSet))
}

func New2(conf *config.Config, orm *gorm.DB, redis *redis.Redis) service.Service {
	panic(wire.Build(providerSet2))
}
