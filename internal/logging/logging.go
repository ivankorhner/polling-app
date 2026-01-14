package logging

import (
	"log/slog"
	"os"
)

func NewLogger(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}

	return slog.New(ContextHandler{Handler: slog.NewJSONHandler(os.Stdout, opts)})
}
