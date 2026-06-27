package observability

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	SERVICE_NAME = "hookwave-server"
)

type Logger struct {
	*zap.Logger
}

func NewLogger() *Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.CallerKey = "logging-at"
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(),
		},
	}
	logger := zap.Must(config.Build()).With(NewTag("service", SERVICE_NAME))

	hookwaveLogger := &Logger{
		Logger: logger,
	}

	return hookwaveLogger

}

func NewTag(key string, value any) zap.Field {
	switch value := value.(type) {
	case int:
		return zap.Int(key, value)
	case int32:
		return zap.Int32(key, value)
	case int64:
		return zap.Int64(key, value)
	case string:
		return zap.String(key, value)
	case time.Duration:
		return zap.Duration(key, value)
	default:
		return zap.Any(key, value)
	}

}

func NewTagError(err error) zap.Field {
	return zap.Error(err)
}
