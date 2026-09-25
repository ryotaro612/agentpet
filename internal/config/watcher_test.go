package config

import (
	"context"
	"io"
	"log/slog"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestWatcherTerminatesOnCancel(t *testing.T) {
	t.Parallel()

	ch, _, cancel := newWatcher(t)
	cancel()

	select {
	case cfg, ok := <-ch:
		if ok {
			t.Fatalf("expected closed channel, got config: %+v", cfg)
		}
		if !reflect.DeepEqual(cfg, Config{}) {
			t.Errorf("a zero-value Config is received from a closed channel: got %+v, want %+v", cfg, Config{})
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("channel was not closed within timeout after cancellation")
	}
}

func TestWatcherCallbackOnFileChange(t *testing.T) {
	t.Parallel()

	ch, path, _ := newWatcher(t)

	const modifiedTOML = `port = 9090

[[pets]]
name = "dog"

[pets.frame]
height = 64
width  = 64

[[pets.animations]]
name     = "run"
filepath = "run.png"
`
	if err := os.WriteFile(path, []byte(modifiedTOML), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-ch:
		if got.Port != 9090 {
			t.Errorf("port: got %d, want 9090", got.Port)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("config was not sent within timeout")
	}
}

const minimalValidTOML = `[[pets]]
name = "cat"

[pets.frame]
height = 32
width  = 32

[[pets.animations]]
name     = "idle"
filepath = "idle.png"
`

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.toml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func newWatcher(t *testing.T) (chan Config, string, context.CancelFunc) {
	t.Helper()
	path := writeConfigFile(t, minimalValidTOML)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	ch := make(chan Config, 1)
	if _, err := NewConfigWatcher(ctx, path, ch, logger); err != nil {
		t.Fatal(err)
	}
	return ch, path, cancel
}
