package config

import (
	"os"

	"go.uber.org/zap"
)

func SetupLogger() {
	var logger *zap.Logger
	var err error

	goEnv := os.Getenv("GO_ENV")

	logger, err = zap.NewProduction()
	if goEnv == "" || goEnv == "development" {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic("failed to initialize zap logger: " + err.Error())
	}

	zap.ReplaceGlobals(logger)
}
