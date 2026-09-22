package internal

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTerminateShutdownsServer(t *testing.T) {
	t.Parallel()

	k, err := NewKeeper("testdata/TestTerminateShutdownsServer.toml")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	s := NewServer(k, NewView(), 0, logger, cancel)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	go s.mcpServer.Run(ctx, serverTransport)

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.1"}, nil)
	cs, err := client.Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}

	cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "terminate"})

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("context was not cancelled after calling terminate")
	}
}
