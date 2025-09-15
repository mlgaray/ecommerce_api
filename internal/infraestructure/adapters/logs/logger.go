package logs

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *logrus.Logger

type contextKey string

const loggerKey contextKey = "logger"

func Init() {
	globalLogger = logrus.New()
	globalLogger.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   "./logs/pps-api.log",
		MaxSize:    10,   // Max size in MB
		MaxBackups: 3,    // Max number of old log files to keep
		MaxAge:     28,   // Max age in days to keep a log file
		Compress:   true, // Compress old log files
	}))

	fmt.Println("Successfully initialized global logger!")
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
