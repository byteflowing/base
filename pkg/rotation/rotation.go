package rotation

import (
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewRotation(opts *configv1.RotationConfig) *lumberjack.Logger {
	if opts == nil {
		panic("Config must not be nil")
	}
	return &lumberjack.Logger{
		Filename:   opts.LogFile,
		MaxSize:    int(opts.MaxSize),
		MaxAge:     int(opts.MaxAge),
		MaxBackups: int(opts.MaxBackups),
		LocalTime:  opts.LocalTime,
		Compress:   opts.Compress,
	}
}
