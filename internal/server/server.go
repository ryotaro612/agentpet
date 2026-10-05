package server

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
	"github.com/ryotaro612/agentpet/internal/pet"
	"github.com/ryotaro612/agentpet/internal/view"
)

type server struct {
	cfgCh     <-chan config.Config
	v         *view.View
	port      int // configured port (0 = auto); updated to the bound port after Listen
	cfgPath   string
	mcpServer *mcp.Server
	logger    *slog.Logger
	cancel    context.CancelFunc
	mu        sync.RWMutex
	k         pet.Keeper
}

func NewServer(cfgCh <-chan config.Config, v *view.View, port int, cfg config.Config, cfgPath string, logger *slog.Logger, cancel context.CancelFunc) (*server, error) {
	if v == nil {
		return nil, fmt.Errorf("server: view is required")
	}
	k := pet.NewKeeper(cfg)

	s := server{
		cfgCh:   cfgCh,
		v:       v,
		port:    port,
		cfgPath: cfgPath,
		mcpServer: mcp.NewServer(&mcp.Implementation{Name: "agentpet", Version: "v0.0.1"}, &mcp.ServerOptions{
			Capabilities: &mcp.ServerCapabilities{
				Tools: &mcp.ToolCapabilities{ListChanged: true},
			},
		}),
		logger: logger,
		cancel: cancel,
		k:      k,
	}

	for _, t := range []mcpTool{
		showWindowTool(s.v),
		hideWindowTool(s.v),
		vetTool(s.cfgPath),
		shuffleAnimationTool(&s),
	} {
		mcp.AddTool(s.mcpServer, &t.tool, t.handler)
	}
	petName := k.DefaultPet()
	s.registerTools(petName, nil)
	if anim, ok := k.DefaultAnim(petName); ok {
		s.v.Show(petName, anim)
	} else {
		s.logger.Warn("no default animation found", "pet", petName)
	}
	return &s, nil
}

func (s *server) onConfigChange(cfg config.Config) {
	k := pet.NewKeeper(cfg)
	petName := k.DefaultPet()
	anim, hasAnim := k.DefaultAnim(petName)
	s.mu.Lock()
	currentPet := s.v.CurrentPet()
	oldAnimTools := playAnimationTools(currentPet, s).names()
	s.k = k
	s.registerTools(petName, oldAnimTools)
	s.mu.Unlock()
	if hasAnim {
		s.v.Show(petName, anim)
	} else {
		s.logger.Warn("no default animation found", "pet", petName)
	}
}

func (s *server) Run(ctx context.Context) error {
	defer s.cancel()

	mux := http.NewServeMux()
	mcpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return s.mcpServer }, nil)
	mux.Handle("/", mcpHandler)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	for {
		ln, err := s.availablePort()
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
		s.logger.Info("MCP server listening", "addr", ln.Addr().String())

		httpSrv := &http.Server{Handler: mux}
		srvDone := make(chan struct{})
		go func() {
			defer close(srvDone)
			if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
				s.logger.Error("server error", "err", err)
			}
		}()

		if !s.serveHTTP(ctx, httpSrv, srvDone, sigCh) {
			return nil
		}
	}
}

// serveHTTP handles events for one HTTP server instance. It returns true when
// the caller should restart with a new listener (port change), false to stop.
func (s *server) serveHTTP(ctx context.Context, httpSrv *http.Server, srvDone <-chan struct{}, sigCh <-chan os.Signal) bool {
	for {
		select {
		case <-ctx.Done():
			httpSrv.Shutdown(context.Background())
			<-srvDone
			return false
		case <-sigCh:
			s.cancel()
			httpSrv.Shutdown(context.Background())
			<-srvDone
			return false
		case cfg, ok := <-s.cfgCh:
			if !ok {
				httpSrv.Shutdown(context.Background())
				<-srvDone
				return false
			}
			s.mu.RLock()
			activePort := s.port
			s.mu.RUnlock()
			s.onConfigChange(cfg)
			if cfg.Port != 0 && cfg.Port != activePort {
				httpSrv.Shutdown(context.Background())
				<-srvDone
				s.mu.Lock()
				s.port = cfg.Port
				s.mu.Unlock()
				s.logger.Info("MCP server restarting", "port", cfg.Port)
				return true
			}
		}
	}
}

func (s *server) availablePort() (net.Listener, error) {
	s.mu.RLock()
	port := s.port
	s.mu.RUnlock()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.port = ln.Addr().(*net.TCPAddr).Port
	s.mu.Unlock()
	return ln, nil
}
