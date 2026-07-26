package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger provides logging functionality with context support
type Logger struct {
	logger *slog.Logger
}

// NewLogger creates a new Logger instance
func NewLogger() *Logger {
	return &Logger{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}
}

// Get returns the underlying logger for compatibility with slog patterns
func (l *Logger) Get() *slog.Logger {
	return l.logger
}

// Log writes a log entry at the specified level
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, keysAndValues ...interface{}) {
	l.logger.Log(ctx, level, msg, keysAndValues...)
}

// Debug logs a debug-level message
func (l *Logger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, slog.LevelDebug, msg, keysAndValues...)
}

// Info logs an info-level message
func (l *Logger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, slog.LevelInfo, msg, keysAndValues...)
}

// Warn logs a warning-level message
func (l *Logger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, slog.LevelWarn, msg, keysAndValues...)
}

// Error logs an error-level message
func (l *Logger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.log(ctx, slog.LevelError, msg, keysAndValues...)
}

// LogV logs a verbose-level message with formatted string
func (l *Logger) Logf(ctx context.Context, format string, v ...interface{}) {
	l.logger.Log(ctx, slog.LevelDebug, format, v...)
}

// Debugf logs a debug-level message with formatted string
func (l *Logger) Debugf(ctx context.Context, format string, v ...interface{}) {
	l.logger.Log(ctx, slog.LevelDebug, format, v...)
}

// Infof logs an info-level message with formatted string
func (l *Logger) Infof(ctx context.Context, format string, v ...interface{}) {
	l.logger.Log(ctx, slog.LevelInfo, format, v...)
}

// Warnf logs a warning-level message with formatted string
func (l *Logger) Warnf(ctx context.Context, format string, v ...interface{}) {
	l.logger.Log(ctx, slog.LevelWarn, format, v...)
}

// Errorf logs an error-level message with formatted string
func (l *Logger) Errorf(ctx context.Context, format string, v ...interface{}) {
	l.logger.Log(ctx, slog.LevelError, format, v...)
}

// Internal logging function
func (l *Logger) log(ctx context.Context, level slog.Level, msg string, keysAndValues ...interface{}) {
	l.logger.Log(ctx, level, msg, keysAndValues...)
}
