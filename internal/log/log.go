package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger for consistent logging throughout the application
type Logger struct {
	*zap.Logger
}

// New creates a new structured logger with the specified log level
func New(logLevel string) (*Logger, error) {
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.DisableCaller = false
	config.DisableStacktrace = level != zapcore.DebugLevel

	logger, err := config.Build(
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Logger{logger}, nil
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.With(fields...)}
}

// Close flushes any buffered log entries
func (l *Logger) Close() error {
	return l.Sync()
}
