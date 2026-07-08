package lexer

import "log/slog"

func ScanToken() {
	// This log will only appear if the CLI has initialized the logger to LevelDebug
	slog.Debug("Generated token", "type", "CHARACTER", "value", "Romeo")
}