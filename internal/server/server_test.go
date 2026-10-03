package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/ryotaro612/agentpet/internal/config"
	"github.com/ryotaro612/agentpet/internal/view"
)

func newTestView(t *testing.T) *view.View {
	t.Helper()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	v, err := view.New(view.NoopInternalView{}, logger)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func catConfig() config.Config {
	return config.Config{
		Pet: "cat",
		Pets: []config.PetConfig{{
			Name:  "cat",
			Frame: config.DimensionConfig{Width: 32, Height: 32},
			Animations: []config.AnimationConfig{
				{Name: "idle", File: "idle.png"},
			},
		}},
	}
}

func dogConfig() config.Config {
	return config.Config{
		Pet: "dog",
		Pets: []config.PetConfig{{
			Name:  "dog",
			Frame: config.DimensionConfig{Width: 32, Height: 32},
			Animations: []config.AnimationConfig{
				{Name: "walk", File: "walk.png"},
			},
		}},
	}
}

// freePort returns an available TCP port by binding then immediately closing.
func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

// waitForTCP polls until something is listening on port or the timeout elapses.
func waitForTCP(t *testing.T, port int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("nothing listening on port %d after %s", port, timeout)
}

func TestServerConfigChange(t *testing.T) {
	t.Parallel()

	t.Run("restarts on the port specified in a new configuration", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
		cfgCh := make(chan config.Config, 1)
		v := newTestView(t)

		initialPort := freePort(t)
		s, err := NewServer(cfgCh, v, initialPort, catConfig(), "", logger, cancel)
		if err != nil {
			t.Fatal(err)
		}
		go s.Run(ctx)

		waitForTCP(t, initialPort, 5*time.Second)

		newPort := freePort(t)
		newCfg := catConfig()
		newCfg.Port = newPort
		cfgCh <- newCfg

		waitForTCP(t, newPort, 5*time.Second)

		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", initialPort), 200*time.Millisecond)
		if err == nil {
			conn.Close()
			t.Errorf("server is still accepting connections on old port %d after restart on port %d", initialPort, newPort)
		}
	})

	t.Run("passes the default pet and animation from a new configuration to the view", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
		cfgCh := make(chan config.Config, 1)
		v := newTestView(t)

		s, err := NewServer(cfgCh, v, 0, catConfig(), "", logger, cancel)
		if err != nil {
			t.Fatal(err)
		}
		go s.Run(ctx)

		cfgCh <- dogConfig()

		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) && v.CurrentPet() != "dog" {
			time.Sleep(10 * time.Millisecond)
		}

		if got := v.CurrentPet(); got != "dog" {
			t.Errorf("view shows pet %q after config change, want %q", got, "dog")
		}
		if got := v.CurrentAnimation(); got != "walk" {
			t.Errorf("view shows animation %q after config change, want %q", got, "walk")
		}
	})
}
