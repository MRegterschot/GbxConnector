package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func SetupLogger() {
	var logger *zap.Logger

	core := zapcore.NewCore(
		setupEncoder(),
		zapcore.AddSync(zapcore.Lock(os.Stdout)),
		getLogLevel(),
	)

	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.ReplaceGlobals(logger)
}

func setupEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.TimeKey = "time"
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewConsoleEncoder(config)
}

func getLogLevel() zapcore.Level {
	switch AppEnv.LogLevel {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel // Default level
	}
}
