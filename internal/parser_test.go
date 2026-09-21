package internal

import (
	"log/slog"
	"testing"
)

func TestLogLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		verbose bool
		want    slog.Level
	}{
		{"returns debug level when verbose is enabled", true, slog.LevelDebug},
		{"returns info level when verbose is disabled", false, slog.LevelInfo},
	}
	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			a := args{verbose: c.verbose}
			if got := a.LogLevel(); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}
