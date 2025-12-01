package logx

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/byteflowing/base/pkg/rotation"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
)

const (
	defaultNameKey = "SRV"
)

func newZap(config *configv1.ZapLogConfig) *zap.Logger {
	var logger *zap.Logger
	opts := getOptions(config)
	if len(config.Outputs) == 0 {
		cfg := getConfig(config)
		var err error
		logger, err = cfg.Build(opts...)
		if err != nil {
			panic(err)
		}
	} else {
		cfg := getEncoderConfig(config)
		cores := getCores(config, cfg)
		logger = zap.New(zapcore.NewTee(cores...), opts...)
	}
	if config.ServiceName != "" {
		logger = logger.Named(config.ServiceName)
	}
	return logger
}

func getOptions(config *configv1.ZapLogConfig) []zap.Option {
	var opts []zap.Option
	opts = append(opts, zap.WithCaller(config.ReportCaller))
	opts = append(opts, zap.AddCallerSkip(int(config.CallerSkip)))
	opts = append(opts, zap.AddStacktrace(zap.NewAtomicLevelAt(convertLogLevel(config.AddStackTraceLevel))))
	return opts
}

func convertLogLevel(l enumv1.LogLevel) zapcore.Level {
	switch l {
	case enumv1.LogLevel_LOG_LEVEL_DEBUG:
		return zapcore.DebugLevel
	case enumv1.LogLevel_LOG_LEVEL_INFO:
		return zapcore.InfoLevel
	case enumv1.LogLevel_LOG_LEVEL_WARN:
		return zapcore.WarnLevel
	case enumv1.LogLevel_LOG_LEVEL_ERROR:
		return zapcore.ErrorLevel
	case enumv1.LogLevel_LOG_LEVEL_PANIC:
		return zapcore.PanicLevel
	case enumv1.LogLevel_LOG_LEVEL_FATAL:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func getLogLevels(lvls []enumv1.LogLevel) []zapcore.Level {
	var levels []zapcore.Level
	for _, lvl := range lvls {
		levels = append(levels, convertLogLevel(lvl))
	}
	return levels
}

func getConfig(config *configv1.ZapLogConfig) zap.Config {
	var cfg zap.Config
	if config.Mode == enumv1.LogMode_LOG_MODE_DEV {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	encoderCfg := getEncoderConfig(config)
	cfg.Level.SetLevel(convertLogLevel(config.Level))
	cfg.EncoderConfig = encoderCfg
	switch config.Format {
	case enumv1.LogFormat_LOG_FORMAT_CONSOLE:
		cfg.Encoding = "console"
	case enumv1.LogFormat_LOG_FORMAT_JSON:
		cfg.Encoding = "json"
	}
	return cfg
}

func getEncoderConfig(config *configv1.ZapLogConfig) zapcore.EncoderConfig {
	var cfg zapcore.EncoderConfig
	if config.Mode == enumv1.LogMode_LOG_MODE_DEV {
		cfg = zap.NewDevelopmentEncoderConfig()
	} else {
		cfg = zap.NewProductionEncoderConfig()
	}
	if config.ShortCaller {
		cfg.EncodeCaller = zapcore.ShortCallerEncoder
	} else {
		cfg.EncodeCaller = zapcore.FullCallerEncoder
	}
	cfg.NameKey = defaultNameKey
	if config.Keys != nil {
		cfg.LevelKey = config.Keys.LevelKey
		cfg.MessageKey = config.Keys.MessageKey
		cfg.TimeKey = config.Keys.TimeKey
		cfg.CallerKey = config.Keys.CallerKey
		cfg.StacktraceKey = config.Keys.StackTraceKey
		cfg.FunctionKey = config.Keys.FunctionKey
	}
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	return cfg
}

func getCores(c *configv1.ZapLogConfig, enc zapcore.EncoderConfig) []zapcore.Core {
	var cores []zapcore.Core
	globalLevel := convertLogLevel(c.Level)
	for _, output := range c.Outputs {
		lvls := getLogLevels(output.Levels)
		core := zapcore.NewCore(
			getEncoders(c, enc),
			zapcore.AddSync(getOutput(output)),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= globalLevel && levelEnablerFunc(lvls)(lvl)
			}),
		)
		cores = append(cores, core)
	}
	return cores
}

func levelEnablerFunc(levels []zapcore.Level) func(lvl zapcore.Level) bool {
	set := make(map[zapcore.Level]struct{}, len(levels))
	for _, lvl := range levels {
		set[lvl] = struct{}{}
	}
	return zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		_, ok := set[lvl]
		return ok
	})
}

func getEncoders(c *configv1.ZapLogConfig, enc zapcore.EncoderConfig) zapcore.Encoder {
	switch c.Format {
	case enumv1.LogFormat_LOG_FORMAT_JSON:
		return zapcore.NewJSONEncoder(enc)
	case enumv1.LogFormat_LOG_FORMAT_CONSOLE:
		return zapcore.NewConsoleEncoder(enc)
	default:
		panic("unknown formatter: " + c.Format.String())
	}
}

func getOutput(out *configv1.ZapLogOutput) io.Writer {
	switch out.Output {
	case enumv1.LogOut_LOG_OUT_STDOUT:
		return os.Stdout
	case enumv1.LogOut_LOG_OUT_STDERR:
		return os.Stderr
	case enumv1.LogOut_LOG_OUT_FILE:
		return rotation.NewRotation(out.LogFile)
	default:
		panic("unknown output format: " + out.Output.String())
	}
}
