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
	keeper    *keeper
	view      view
	portNum   int
	mcpServer *mcp.Server
	handler   http.Handler
	logger    *slog.Logger
	cancel    context.CancelFunc
}

func NewServer(k *keeper, view view, port int, logger *slog.Logger, cancel context.CancelFunc) server {
	s := server{
		keeper:    k,
		view:      view,
		portNum:   port,
		mcpServer: mcp.NewServer(&mcp.Implementation{Name: "agentpet", Version: "v0.0.1"}, nil),
		logger:    logger,
		cancel:    cancel,
	}
	addTerminate := func() {
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        "terminate",
			Description: "Terminate the agentpet MCP server",
		}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			defer cancel()
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "Server is shutting down"}},
			}, nil, nil
		})
	}
	addTerminate()
	mcpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return s.mcpServer }, nil)
	s.handler = configReloadMiddleware(k, addTerminate, logger, mcpHandler)
	return s
}

func (s server) Run(ctx context.Context) {
	defer s.cancel()

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

	httpSrv := &http.Server{Handler: s.handler}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		select {
		case <-ch:
			s.cancel()
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

func configReloadMiddleware(k *keeper, notify func(), logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		changed, err := k.reloadIfUpdated()
		if err != nil {
			logger.Warn("config reload failed", "err", err)
		} else if changed {
			notify()
		}
		next.ServeHTTP(w, r)
	})
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
