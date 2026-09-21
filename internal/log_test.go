package internal

import (
	"context"
	"log/slog"
	"testing"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		level       slog.Level
		probe       slog.Level
		wantEnabled bool
	}{
		{"enables debug messages when created with debug level", slog.LevelDebug, slog.LevelDebug, true},
		{"suppresses debug messages when created with info level", slog.LevelInfo, slog.LevelDebug, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			logger := NewLogger(c.level)
			got := logger.Handler().Enabled(context.Background(), c.probe)
			if got != c.wantEnabled {
				t.Errorf("Enabled(%v) = %v, want %v", c.probe, got, c.wantEnabled)
			}
		})
	}
}
