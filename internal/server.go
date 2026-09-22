package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type server struct {
	keeper    keeper
	view      view
	portNum   int
	mcpServer *mcp.Server
	logger    *slog.Logger
}

func NewServer(keeper keeper, view view, port int, logger *slog.Logger) server {
	return server{
		keeper:    keeper,
		view:      view,
		portNum:   port,
		mcpServer: mcp.NewServer(&mcp.Implementation{Name: "agentpet", Version: "v0.0.1"}, nil),
		logger:    logger,
	}
}

func (s server) Run(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "terminate",
		Description: "Terminate the agentpet MCP server",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		defer cancel()
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Server is shutting down"}},
		}, nil, nil
	})

	port, err := s.availablePort()
	if err != nil {
		s.logger.Error("failed to find available port", "err", err)
		return
	}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		s.logger.Error("failed to listen", "err", err)
		return
	}
	s.logger.Info("MCP server listening", "addr", ln.Addr().String())

	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return s.mcpServer }, nil)
	httpSrv := &http.Server{Handler: handler}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()

	go func() {
		<-ctx.Done()
		httpSrv.Shutdown(context.Background())
	}()

	if err := httpSrv.Serve(ln); err != http.ErrServerClosed {
		s.logger.Error("server error", "err", err)
	}
}

func (s server) availablePort() (int, error) {
	if s.portNum != 0 {
		return s.portNum, nil
	}
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
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
