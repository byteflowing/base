//go:build wireinject
// +build wireinject

package message

import (
	"github.com/google/wire"

	"github.com/byteflowing/base/app/message/service"
	"github.com/byteflowing/base/singleton"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
)

var providerSet = wire.NewSet(
	singleton.NewRDB,
	service.NewMessageService,
	wire.FieldsOf(new(*configv1.Config), "Redis"),
)

func NewMessageServiceWithConfig(cfg *configv1.Config) *service.MessageService {
	panic(wire.Build(providerSet))
}
