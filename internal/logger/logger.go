package logger

import (
	"log/slog"
	"os"
)

// Logger is the global logger instance
var Logger *slog.Logger

func init() {
	// Create a new logger instance
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
