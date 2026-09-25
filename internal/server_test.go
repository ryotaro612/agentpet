package internal

import (
	"context"
	"io"
	"log/slog"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal/config"
)

func TestToolList(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cfg, err := config.LoadConfig("testdata/TestToolList.toml")
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	cfgCh := make(chan config.Config)
	if _, err := config.NewConfigWatcher(ctx, "testdata/TestToolList.toml", cfgCh, logger); err != nil {
		t.Fatal(err)
	}

	s := NewServer(cfgCh, nil, 0, cfg, logger, cancel)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	go s.mcpServer.Run(ctx, serverTransport)

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cs, err := client.Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}

	result, err := cs.ListTools(t.Context(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatal(err)
	}

	got := make([]string, len(result.Tools))
	for i, tool := range result.Tools {
		got[i] = tool.Name
	}
	sort.Strings(got)

	want := []string{"anim_idle", "anim_walk", "hide_window", "list_pets", "show_window"}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("tool list mismatch (-want +got):\n%s", diff)
	}
}
