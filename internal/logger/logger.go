package logger

import (
	"os"
	"log/slog"
)

// Logger is the global logger instance
var Logger *slog.Logger

func init() {
	// Create a new logger instance
	Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}