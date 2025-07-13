package config

import (
	"github.com/byteflowing/go-common/rotation"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/tool/internal_pkg/log"
)

type LogConfig struct {
	// 日志级别
	// trace
	// debug
	// info
	// notice
	// warn
	// error
	// fatal
	Level string
	// 如果配置此项日志将写入文件中
	Output *rotation.Config
}

func (l *LogConfig) LogLevel() klog.Level {
	switch l.Level {
	case "trace":
		return klog.LevelTrace
	case "debug":
		return klog.LevelDebug
	case "info":
		return klog.LevelInfo
	case "notice":
		return klog.LevelNotice
	case "warn":
		return klog.LevelWarn
	case "error":
		return klog.LevelError
	case "fatal":
		return klog.LevelFatal
	}
	log.Warnf("unknown log level: %s, log level will set to level info", l.Level)
	return klog.LevelInfo
}
