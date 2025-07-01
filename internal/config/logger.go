package config

import (
	"sea-catering-backend/pkg/logger"
)

func NewLogger() *logger.Logger {
	config := logger.LoadConfigFromEnv()
	return logger.New(config)
}

func NewLoggerWithConfig(config *logger.Config) *logger.Logger {
	return logger.New(config)
}
