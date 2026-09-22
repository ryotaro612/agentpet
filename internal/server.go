package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type server struct {
	keeper         keeper
	view           view
	port           int
	mcpServer      *mcp.Server
	configFilePath string
	logger         *slog.Logger
}

func NewServer(keeper keeper, view view, port int, configFilePath string) server {

	panic("not implemented")
}

func (s server) Run() {
	panic("not implemented")
}

func RunServer(ctx context.Context, logger *slog.Logger, cfg Config, cancel context.CancelFunc) error {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "agentpet",
		Version: "v0.0.1",
	}, nil)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "terminate",
		Description: "Terminate the agentpet MCP server",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		defer cancel()
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Server is shutting down"}},
		}, nil, nil
	})

	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return s }, nil)

	port := cfg.Server.Port
	if port == 0 {
		port = 8080
	}
	addr := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	logger.Info("MCP server listening", "addr", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
