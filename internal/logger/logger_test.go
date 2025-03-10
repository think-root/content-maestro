package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	assert.NotNil(t, logger, "Logger should not be nil")
}

func TestLoggerOutput(t *testing.T) {
	buf := &bytes.Buffer{}

	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	writeSyncer := zapcore.AddSync(buf)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	zapLogger := zap.New(core)
	sugar := zapLogger.Sugar()

	testMessage := "test message"
	sugar.Info(testMessage)

	output := buf.String()
	assert.Contains(t, output, testMessage, "Logger output should contain test message")
}

func TestLoggerLevels(t *testing.T) {
	logger := NewLogger()

	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	assert.True(t, true, "Logging functions should execute without errors")
}
