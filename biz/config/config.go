package config

import (
	"github.com/byteflowing/go-common/config"
	"github.com/byteflowing/go-common/orm"
	"github.com/byteflowing/go-common/redis"
)

type Config struct {
	HostPort string         // 服务监听地址 e.g. 0.0.0.0:8888
	Log      *LogConfig     // 日志配置
	DB       *orm.Config    // 数据库配置
	RDB      *redis.Config  // redis配置
	Message  *MessageConfig // message模块配置
	User     *UserConfig    //  user模块配置
}

func New(configFile string) (conf *Config) {
	c := &Config{}
	if err := config.ReadConfig(configFile, c); err != nil {
		panic(err)
	}
	return c
}
