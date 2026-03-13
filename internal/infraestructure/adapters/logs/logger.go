package logs

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var globalLogger *logrus.Logger

type contextKey string

const loggerKey contextKey = "logger"

func Init() {
	globalLogger = logrus.New()
	globalLogger.SetLevel(resolveLogLevel())
	globalLogger.SetOutput(io.MultiWriter(os.Stdout))

	fmt.Printf("Successfully initialized global logger! Level: %s\n", globalLogger.GetLevel())
}

func resolveLogLevel() logrus.Level {
	// LOG_LEVEL tiene prioridad si está definido (permite override explícito)
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		parsed, err := logrus.ParseLevel(level)
		if err == nil {
			return parsed
		}
		fmt.Printf("Invalid LOG_LEVEL '%s', falling back to environment-based level\n", level)
	}

	// Si no hay LOG_LEVEL, se determina por ENVIRONMENT
	switch os.Getenv("ENVIRONMENT") {
	case "production":
		return logrus.ErrorLevel
	case "test":
		return logrus.WarnLevel
	default: // develop, development, o vacío
		return logrus.DebugLevel
	}
}

func WithFields(fields map[string]interface{}) *logrus.Entry {
	return globalLogger.WithFields(logrus.Fields(fields))
}

func SetLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *logrus.Entry {
	if logger, ok := ctx.Value(loggerKey).(*logrus.Entry); ok {
		return logger
	}
	return globalLogger.WithContext(ctx)
}

func Error(args ...interface{}) {
	globalLogger.Error(args...)
}

func Info(args ...interface{}) {
	globalLogger.Info(args...)
}

func Warn(args ...interface{}) {
	globalLogger.Warn(args...)
}

func Debug(args ...interface{}) {
	globalLogger.Debug(args...)
}
