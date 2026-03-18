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

	// JSON formatter for all environments except development (human-readable text in local dev)
	// test + production emit JSON so Promtail/Loki can parse structured fields
	env := os.Getenv("ENVIRONMENT")
	if env != "development" && env != "" {
		globalLogger.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime: "timestamp",
				logrus.FieldKeyMsg:  "message",
			},
		})
	}

	// Add default fields to every log entry via a hook
	globalLogger.AddHook(&defaultFieldsHook{
		fields: logrus.Fields{
			"service":     "ecommerce-api",
			"environment": env,
		},
	})

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

// defaultFieldsHook injects base fields into every log entry.
type defaultFieldsHook struct {
	fields logrus.Fields
}

func (h *defaultFieldsHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *defaultFieldsHook) Fire(entry *logrus.Entry) error {
	for k, v := range h.fields {
		if _, exists := entry.Data[k]; !exists {
			entry.Data[k] = v
		}
	}
	return nil
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
