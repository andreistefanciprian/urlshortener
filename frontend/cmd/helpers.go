package main

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// getEnvAsIntOrDefault returns environment variable as int or default if not set or invalid
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initLogger() *logrus.Logger {
	logger := logrus.New()
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		PadLevelText:  true,
	})
	logger.Infof("Logger initialized with log level: %s", logLevel)
	return logger
}
