//go:build wireinject
// +build wireinject

package user

import (
	"github.com/byteflowing/base/dal"
	"github.com/byteflowing/base/gen/config/v1"
	"github.com/byteflowing/base/pkg/captcha"
	"github.com/byteflowing/base/pkg/common"
	"github.com/byteflowing/base/pkg/msg/mail"
	"github.com/byteflowing/base/pkg/msg/sms"
	"github.com/google/wire"
)

var publicSet = wire.NewSet(
	common.NewDb,
	common.NewRDB,
	common.NewDistributedLock,
	common.NewGlobalIdGenerator,
	common.NewShortIDGenerator,
	common.NewWechatManager,
	captcha.NewSmsCaptcha,
	captcha.NewMailCaptcha,
	captcha.NewCaptcha,
	sms.New,
	mail.New,
	dal.New,
)

var userProviderSet = wire.NewSet(
	NewCache,
	NewRepo,
	NewJwtService,
	NewTwoStepVerifier,
	NewAuthLimiter,
	New,
	NewSessionBlockList,
	NewConfig,
	wire.FieldsOf(new(*configv1.Config), "Sms", "Mail", "Captcha", "GlobalId", "ShortId", "Wechat", "Db", "Redis", "DistributedLock", "User"),
	wire.FieldsOf(new(*configv1.User), "AuthLimiter", "Jwt", "TwoStepVerifier", "Cache", "SessionBlockList"),
)

func NewWithConfig(confFile string) *Impl {
	panic(wire.Build(
		publicSet,
		userProviderSet,
	))
}
