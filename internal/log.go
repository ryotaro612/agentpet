package internal

import (
	"log/slog"
	"os"
)

func NewLogger(logLevel slog.Level) *slog.Logger {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	return slog.New(h)
}
