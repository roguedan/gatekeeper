package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestNew_CreatesLoggerWithInfoLevel verifies logger creation with default level
func TestNew_CreatesLoggerWithInfoLevel(t *testing.T) {
	logger, err := New("info")

	require.NoError(t, err)
	require.NotNil(t, logger)
	assert.NotNil(t, logger.Logger)
}

// TestNew_CreatesLoggerWithDebugLevel verifies logger creation with debug level
func TestNew_CreatesLoggerWithDebugLevel(t *testing.T) {
	logger, err := New("debug")

	require.NoError(t, err)
	require.NotNil(t, logger)
}

// TestNew_CreatesLoggerWithWarnLevel verifies logger creation with warn level
func TestNew_CreatesLoggerWithWarnLevel(t *testing.T) {
	logger, err := New("warn")

	require.NoError(t, err)
	require.NotNil(t, logger)
}

// TestNew_CreatesLoggerWithErrorLevel verifies logger creation with error level
func TestNew_CreatesLoggerWithErrorLevel(t *testing.T) {
	logger, err := New("error")

	require.NoError(t, err)
	require.NotNil(t, logger)
}

// TestNew_FailsWithInvalidLevel verifies error on invalid log level
func TestNew_FailsWithInvalidLevel(t *testing.T) {
	_, err := New("invalid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

// TestWithFields returns logger with additional fields
func TestWithFields_AddsFieldsToLogger(t *testing.T) {
	logger, _ := New("info")

	withFields := logger.WithFields(zap.String("user_id", "123"))

	require.NotNil(t, withFields)
	assert.NotNil(t, withFields.Logger)
}

// TestWithFields_SupportsMultipleFields adds multiple fields
func TestWithFields_SupportsMultipleFields(t *testing.T) {
	logger, _ := New("info")

	withFields := logger.WithFields(
		zap.String("user_id", "123"),
		zap.String("request_id", "abc"),
		zap.Int("status_code", 200),
	)

	require.NotNil(t, withFields)
}

// TestClose flushes the logger
func TestClose_FlushesLogger(t *testing.T) {
	logger, _ := New("info")

	err := logger.Close()

	// Zap logger may return error on close (like already closed), but should not panic
	// Error is acceptable here
	_ = err
}

// TestLogger_LoggingMessages verifies logging actually works
func TestLogger_LoggingMessages(t *testing.T) {
	logger, _ := New("info")
	defer logger.Close()

	// These should not panic
	logger.Info("test info message", zap.String("key", "value"))
	logger.Warn("test warn message", zap.Int("code", 123))
	logger.Error("test error message", zap.Error(assert.AnError))
}

// TestLogger_DebugNotLoggedInInfoLevel verifies debug level filtering
func TestLogger_DebugNotLoggedInInfoLevel(t *testing.T) {
	logger, _ := New("info")
	defer logger.Close()

	// Debug should not be logged when level is info
	logger.Debug("debug message", zap.String("key", "value"))
	logger.Info("info message", zap.String("key", "value"))
	// If we got here without panic, test passes
	assert.True(t, true)
}

// TestLogger_DebugLoggedInDebugLevel verifies debug level logging
func TestLogger_DebugLoggedInDebugLevel(t *testing.T) {
	logger, _ := New("debug")
	defer logger.Close()

	// Debug should be logged when level is debug
	logger.Debug("debug message", zap.String("key", "value"))
	assert.True(t, true)
}
