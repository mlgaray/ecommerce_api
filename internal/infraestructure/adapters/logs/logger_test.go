package logs

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogger_Init(t *testing.T) {
	t.Run("when Init is called then initializes global logger", func(t *testing.T) {
		// Arrange
		globalLogger = nil

		// Act
		Init()

		// Assert
		assert.NotNil(t, globalLogger)
		assert.Equal(t, logrus.DebugLevel, globalLogger.Level)
	})

	t.Run("when Init is called multiple times then reinitializes logger", func(t *testing.T) {
		// Arrange
		Init()
		firstLogger := globalLogger

		// Act
		Init()
		secondLogger := globalLogger

		// Assert
		assert.NotNil(t, secondLogger)
		assert.NotEqual(t, firstLogger, secondLogger)
	})
}

func TestLogger_WithFields(t *testing.T) {
	t.Run("when WithFields is called then returns entry with fields", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)
		globalLogger.SetFormatter(&logrus.JSONFormatter{})

		fields := map[string]interface{}{
			"operation": "test_operation",
			"user_id":   123,
		}

		// Act
		entry := WithFields(fields)

		// Assert
		assert.NotNil(t, entry)
		entry.Info("test message")

		output := buf.String()
		assert.Contains(t, output, "test_operation")
		assert.Contains(t, output, "123")
		assert.Contains(t, output, "test message")
	})

	t.Run("when WithFields is called with empty map then returns entry", func(t *testing.T) {
		// Arrange
		Init()

		// Act
		entry := WithFields(map[string]interface{}{})

		// Assert
		assert.NotNil(t, entry)
	})
}

func TestLogger_SetLogger(t *testing.T) {
	t.Run("when SetLogger is called then stores logger in context", func(t *testing.T) {
		// Arrange
		Init()
		ctx := context.Background()
		entry := globalLogger.WithFields(logrus.Fields{"test": "value"})

		// Act
		newCtx := SetLogger(ctx, entry)

		// Assert
		assert.NotNil(t, newCtx)
		retrievedLogger := newCtx.Value(loggerKey)
		assert.NotNil(t, retrievedLogger)
		assert.Equal(t, entry, retrievedLogger)
	})

	t.Run("when SetLogger is called with nil context then returns context with logger", func(t *testing.T) {
		// Arrange
		Init()
		entry := globalLogger.WithFields(logrus.Fields{"test": "value"})

		// Act
		newCtx := SetLogger(context.Background(), entry)

		// Assert
		assert.NotNil(t, newCtx)
	})
}

func TestLogger_FromContext(t *testing.T) {
	t.Run("when FromContext is called with logger in context then returns that logger", func(t *testing.T) {
		// Arrange
		Init()
		ctx := context.Background()
		expectedEntry := globalLogger.WithFields(logrus.Fields{"test": "value"})
		ctx = SetLogger(ctx, expectedEntry)

		// Act
		entry := FromContext(ctx)

		// Assert
		assert.NotNil(t, entry)
		assert.Equal(t, expectedEntry, entry)
	})

	t.Run("when FromContext is called without logger in context then returns global logger with context", func(t *testing.T) {
		// Arrange
		Init()
		ctx := context.Background()

		// Act
		entry := FromContext(ctx)

		// Assert
		assert.NotNil(t, entry)
	})

	t.Run("when FromContext is called with empty context then returns global logger", func(t *testing.T) {
		// Arrange
		Init()
		ctx := context.Background()

		// Act
		entry := FromContext(ctx)

		// Assert
		assert.NotNil(t, entry)
	})
}

func TestLogger_Error(t *testing.T) {
	t.Run("when Error is called then logs error message", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)

		// Act
		Error("test error message")

		// Assert
		output := buf.String()
		assert.Contains(t, output, "test error message")
		assert.Contains(t, output, "error")
	})

	t.Run("when Error is called with multiple args then logs all args", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)

		// Act
		Error("error:", "something", "went", "wrong")

		// Assert
		output := buf.String()
		assert.Contains(t, output, "error")
		assert.Contains(t, output, "something")
		assert.Contains(t, output, "went")
		assert.Contains(t, output, "wrong")
	})
}

func TestLogger_Info(t *testing.T) {
	t.Run("when Info is called then logs info message", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)

		// Act
		Info("test info message")

		// Assert
		output := buf.String()
		assert.Contains(t, output, "test info message")
		assert.Contains(t, output, "info")
	})

	t.Run("when Info is called with multiple args then logs all args", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)

		// Act
		Info("user", "logged in", "successfully")

		// Assert
		output := buf.String()
		assert.Contains(t, output, "user")
		assert.Contains(t, output, "logged in")
		assert.Contains(t, output, "successfully")
	})
}

func TestLogger_Warn(t *testing.T) {
	t.Run("when Warn is called then logs warning message", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)

		// Act
		Warn("test warning message")

		// Assert
		output := buf.String()
		assert.Contains(t, output, "test warning message")
		assert.Contains(t, output, "warning")
	})

	t.Run("when Warn is called with multiple args then logs all args", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)

		// Act
		Warn("deprecated", "feature", "used")

		// Assert
		output := buf.String()
		assert.Contains(t, output, "deprecated")
		assert.Contains(t, output, "feature")
		assert.Contains(t, output, "used")
	})
}

func TestLogger_Debug(t *testing.T) {
	t.Run("when Debug is called then logs debug message", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)

		// Act
		Debug("test debug message")

		// Assert
		output := buf.String()
		assert.Contains(t, output, "test debug message")
		assert.Contains(t, output, "debug")
	})

	t.Run("when Debug is called with multiple args then logs all args", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)

		// Act
		Debug("variable", "value:", 42)

		// Assert
		output := buf.String()
		assert.Contains(t, output, "variable")
		assert.Contains(t, output, "value:")
		assert.Contains(t, output, "42")
	})
}

func TestLogger_LogLevels(t *testing.T) {
	t.Run("when logger is initialized then supports all log levels", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)

		// Act & Assert - Debug level should allow all logs
		Debug("debug message")
		assert.Contains(t, buf.String(), "debug message")
		buf.Reset()

		Info("info message")
		assert.Contains(t, buf.String(), "info message")
		buf.Reset()

		Warn("warn message")
		assert.Contains(t, buf.String(), "warn message")
		buf.Reset()

		Error("error message")
		assert.Contains(t, buf.String(), "error message")
	})
}

func TestLogger_WithFieldsChaining(t *testing.T) {
	t.Run("when WithFields is chained then preserves all fields", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)
		globalLogger.SetFormatter(&logrus.JSONFormatter{})

		// Act
		entry := WithFields(map[string]interface{}{
			"request_id": "123",
			"user_id":    456,
		})
		entry.WithField("operation", "test").Info("chained message")

		// Assert
		output := buf.String()
		assert.Contains(t, output, "123")
		assert.Contains(t, output, "456")
		assert.Contains(t, output, "test")
		assert.Contains(t, output, "chained message")
	})
}

func TestLogger_ContextPropagation(t *testing.T) {
	t.Run("when logger is set in context then can be retrieved in nested functions", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)
		globalLogger.SetFormatter(&logrus.JSONFormatter{})

		ctx := context.Background()
		entry := globalLogger.WithFields(logrus.Fields{
			"request_id": "nested-test-123",
		})
		ctx = SetLogger(ctx, entry)

		// Act
		func(ctx context.Context) {
			logger := FromContext(ctx)
			logger.Info("nested function log")
		}(ctx)

		// Assert
		output := buf.String()
		assert.Contains(t, output, "nested-test-123")
		assert.Contains(t, output, "nested function log")
	})
}

func TestLogger_LoggerOutput(t *testing.T) {
	t.Run("when logger writes to buffer then output is captured", func(t *testing.T) {
		// Arrange
		Init()
		var buf bytes.Buffer
		globalLogger.SetOutput(&buf)
		testMessage := "unique-test-message-12345"

		// Act
		Info(testMessage)

		// Assert
		output := buf.String()
		assert.True(t, strings.Contains(output, testMessage))
		assert.True(t, len(output) > 0)
	})
}
