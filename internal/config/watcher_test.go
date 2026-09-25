package config

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestWatcherTerminatesOnCancel(t *testing.T) {
	t.Parallel()

	w, path, cancel := newWatcher(t)
	called := make(chan struct{}, 1)
	if err := w.Watch(func(Config) {
		called <- struct{}{}
	}); err != nil {
		t.Fatal(err)
	}

	cancel()
	time.Sleep(100 * time.Millisecond)

	if err := os.WriteFile(path, []byte(minimalValidTOML), 0644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-called:
		t.Fatal("onChange was called after the context was cancelled")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestWatcherCallbackOnFileChange(t *testing.T) {
	t.Parallel()

	w, path, _ := newWatcher(t)
	updated := make(chan Config, 1)
	if err := w.Watch(func(c Config) {
		select {
		case updated <- c:
		default:
		}
	}); err != nil {
		t.Fatal(err)
	}

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
	case got := <-updated:
		if got.Port != 9090 {
			t.Errorf("port: got %d, want 9090", got.Port)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("onChange was not called within timeout")
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

func newWatcher(t *testing.T) (*Watcher, string, context.CancelFunc) {
	t.Helper()
	path := writeConfigFile(t, minimalValidTOML)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	return NewConfigWatcher(ctx, path, cfg, logger), path, cancel
}
