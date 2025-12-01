package logx

import (
	"context"
	"sync"

	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	"go.uber.org/zap"
)

var (
	std           *zap.Logger
	once          sync.Once
	stdConfig     *StdConfig
	defaultConfig *configv1.ZapLogConfig
)

type StdConfig struct {
	CtxLogIdKey string
	LogIdKey    string
}

func CtxWithLogID(ctx context.Context, logID string) context.Context {
	return context.WithValue(ctx, stdConfig.CtxLogIdKey, logID)
}

func GetCtxLogID(ctx context.Context) string {
	if stdConfig != nil && stdConfig.CtxLogIdKey != "" {
		if v, ok := ctx.Value(stdConfig.CtxLogIdKey).(string); ok {
			return v
		}
	}
	return ""
}

func StdWithLogID(logID string) *zap.Logger {
	return std.With(zap.String(stdConfig.LogIdKey, logID))
}

func WithLogID(logger *zap.Logger, logIdKey, logID string) *zap.Logger {
	return logger.With(zap.String(logIdKey, logIdKey))
}

// Init CallerSkip设置为1刚好可以显示记录日志的那行代码
func Init(config *configv1.ZapLogConfig) {
	once.Do(func() {
		if config == nil {
			config = defaultConfig
		}
		std = newZap(config)
		stdConfig = &StdConfig{
			CtxLogIdKey: config.CtxLogIdKey,
			LogIdKey:    config.LogIdKey,
		}
	})
}

func init() {
	defaultConfig = &configv1.ZapLogConfig{
		Mode:               enumv1.LogMode_LOG_MODE_DEV,
		Format:             enumv1.LogFormat_LOG_FORMAT_CONSOLE,
		Level:              enumv1.LogLevel_LOG_LEVEL_INFO,
		ReportCaller:       true,
		ShortCaller:        true,
		CallerSkip:         1,
		AddStackTraceLevel: enumv1.LogLevel_LOG_LEVEL_ERROR,
	}
	conf := getConfig(defaultConfig)
	opts := getOptions(defaultConfig)
	logger, err := conf.Build(opts...)
	if err != nil {
		panic(err)
	}
	std = logger
	stdConfig = &StdConfig{}
}

func NewZapLogger(config *configv1.ZapLogConfig) *zap.Logger {
	return newZap(config)
}

func GetStdLogger() *zap.Logger {
	return std
}

func Sync() error {
	return std.Sync()
}

func Debug(msg string, fields ...zap.Field) {
	std.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	std.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	std.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	std.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	std.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	std.Fatal(msg, fields...)
}

func CtxDebug(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields)
	std.Debug(msg, fs...)
}

func CtxInfo(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields)
	std.Info(msg, fs...)
}

func CtxWarn(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields)
	std.Warn(msg, fs...)
}

func CtxError(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields)
	std.Error(msg, fs...)
}

func CtxPanic(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields)
	std.Panic(msg, fs...)
}

func CtxFatal(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields)
	std.Fatal(msg, fs...)
}

func addLogIdToFields(ctx context.Context, fields []zap.Field) []zap.Field {
	logId := GetCtxLogID(ctx)
	logIdKey := stdConfig.LogIdKey
	if logId == "" || logIdKey == "" {
		return fields
	}
	newFields := make([]zap.Field, 0, len(fields)+1)
	newFields = append(newFields, zap.String(logIdKey, logId))
	newFields = append(newFields, fields...)
	return newFields
}
