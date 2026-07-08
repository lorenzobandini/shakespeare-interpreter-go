package logger

import (
	"log/slog"
	"os"
)

// LogLevel controls the verbosity of the global logger.
type LogLevel int

const (
	LevelInfo LogLevel = iota // Standard production logs
	LevelDebug                // Detailed development logs (e.g., token extraction)
)

// Init initializes the global slog logger with the requested verbosity.
func Init(level LogLevel) {
	var programLevel slog.Level

	switch level {
	case LevelDebug:
		programLevel = slog.LevelDebug
	default:
		programLevel = slog.LevelInfo
	}

	// Uses a TextHandler for human-readable text logs in the console during development
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: programLevel,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}