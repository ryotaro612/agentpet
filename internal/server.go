package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal/config"
)

type server struct {
	watcher   *config.Watcher
	v         *view
	portNum   int
	mcpServer *mcp.Server
	logger    *slog.Logger
	cancel    context.CancelFunc
	animMu    sync.Mutex
	animNames []string
}

func NewServer(w *config.Watcher, v *view, port int, logger *slog.Logger, cancel context.CancelFunc) *server {
	s := &server{
		watcher: w,
		v:       v,
		portNum: port,
		mcpServer: mcp.NewServer(&mcp.Implementation{Name: "agentpet", Version: "v0.0.1"}, &mcp.ServerOptions{
			Capabilities: &mcp.ServerCapabilities{
				Tools: &mcp.ToolCapabilities{ListChanged: true},
			},
		}),
		logger: logger,
		cancel: cancel,
	}
	s.registerTools(w.Get())
	return s
}

func toolHandlerFor(getConfig func() config.Config) func(fn func(config.Config) (*mcp.CallToolResult, any, error)) func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
	return func(fn func(config.Config) (*mcp.CallToolResult, any, error)) func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
		return func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			return fn(getConfig())
		}
	}
}

func (s *server) registerTools(cfg config.Config) {
	anims := animationsFromConfig(cfg)

	newNames := make([]string, 0, len(anims))
	for _, a := range anims {
		newNames = append(newNames, a.name)
	}

	s.animMu.Lock()
	old := s.animNames
	s.animNames = newNames
	s.animMu.Unlock()

	s.mcpServer.RemoveTools(old...)

	withConfig := toolHandlerFor(s.watcher.Get)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "show_window",
		Description: "Show the pet window",
	}, withConfig(func(_ config.Config) (*mcp.CallToolResult, any, error) {
		if s.v != nil {
			s.v.ShowWindow()
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Window is now visible"}}}, nil, nil
	}))

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "hide_window",
		Description: "Hide the pet window",
	}, withConfig(func(_ config.Config) (*mcp.CallToolResult, any, error) {
		if s.v != nil {
			s.v.HideWindow()
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Window is now hidden"}}}, nil, nil
	}))

	for _, anim := range anims {
		animName := anim.name
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        animName,
			Description: anim.description,
		}, withConfig(func(cfg config.Config) (*mcp.CallToolResult, any, error) {
			if s.v == nil {
				return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Displaying " + animName}}}, nil, nil
			}
			for _, a := range animationsFromConfig(cfg) {
				if a.name == animName {
					s.v.Show(a)
					return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Displaying " + animName}}}, nil, nil
				}
			}
			return nil, nil, fmt.Errorf("animation %q not found", animName)
		}))
	}
}

func (s *server) Run(ctx context.Context) {
	defer s.cancel()

	if err := s.watcher.Watch(s.registerTools); err != nil {
		s.logger.Warn("failed to start config watcher", "err", err)
	}

	if s.v != nil {
		s.v.Show(initialAnimation(s.watcher.Get()))
	}

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

	mcpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return s.mcpServer }, nil)
	httpSrv := &http.Server{Handler: mcpHandler}

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

func (s *server) availablePort() (int, error) {
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
