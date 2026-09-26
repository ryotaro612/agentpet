package server

import (
	"context"
	"encoding/json"
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

type changePetInput struct {
	Name      string `json:"name"`
	Animation string `json:"animation"`
}

type server struct {
	cfgCh     <-chan config.Config
	v         *view.View
	port      int // configured port (0 = auto); updated to the bound port after Listen
	mcpServer *mcp.Server
	logger    *slog.Logger
	cancel    context.CancelFunc
	mu        sync.RWMutex
	k         pet.Keeper
}

func NewServer(cfgCh <-chan config.Config, v *view.View, port int, cfg config.Config, logger *slog.Logger, cancel context.CancelFunc) (*server, error) {
	if v == nil {
		return nil, fmt.Errorf("server: view is required")
	}
	k := pet.NewKeeper(cfg)

	s := server{
		cfgCh: cfgCh,
		v:     v,
		port:  port,
		mcpServer: mcp.NewServer(&mcp.Implementation{Name: "agentpet", Version: "v0.0.1"}, &mcp.ServerOptions{
			Capabilities: &mcp.ServerCapabilities{
				Tools: &mcp.ToolCapabilities{ListChanged: true},
			},
		}),
		logger: logger,
		cancel: cancel,
		k:      k,
	}

	show := showWindowTool(s.v)
	mcp.AddTool(s.mcpServer, &show.tool, show.handler)
	hide := hideWindowTool(s.v)
	mcp.AddTool(s.mcpServer, &hide.tool, hide.handler)
	petName := k.DefaultPet()
	anim, _ := k.DefaultAnim(petName)
	s.registerTools(petName)
	s.v.Show(petName, anim)
	return &s, nil
}

func buildChangePetSchema(petNames []string) json.RawMessage {
	b, _ := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Name of the pet to switch to",
				"enum":        petNames,
			},
			"animation": map[string]any{
				"type":        "string",
				"description": "Name of the animation to activate on the new pet",
			},
		},
		"required": []string{"name"},
	})
	return json.RawMessage(b)
}

func (s *server) animToolNames(petName string) []string {
	anims := s.k.Animations(petName)
	names := make([]string, len(anims))
	for i, a := range anims {
		names[i] = "play_" + a.Name
	}
	return names
}

// registerTools must be called with s.mu held for writing.
func (s *server) registerTools(petName string) {
	s.mcpServer.RemoveTools(append(s.animToolNames(petName), "change_pet", "list_pets")...)
	lp := listPetsTool(s.k, s.v)
	mcp.AddTool(s.mcpServer, &lp.tool, lp.handler)

	for _, a := range s.k.Animations(petName) {
		toolName := "play_" + a.Name
		desc := a.Description
		if desc == "" {
			desc = "Play the " + a.Name + " animation"
		}
		animName := a.Name
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        toolName,
			Description: desc,
		}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			currentPet, _ := s.v.Current()
			s.mu.RLock()
			k := s.k
			s.mu.RUnlock()
			anim, err := k.PlayAnimation(currentPet, animName)
			if err != nil {
				return nil, nil, err
			}
			s.v.Show(currentPet, anim)
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Playing " + animName}}}, nil, nil
		})
	}

	if otherPets := s.k.OtherPetNames(petName); len(otherPets) > 0 {
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        "change_pet",
			Description: "Switch the active pet",
			InputSchema: buildChangePetSchema(otherPets),
		}, func(_ context.Context, _ *mcp.CallToolRequest, input changePetInput) (*mcp.CallToolResult, any, error) {
			s.mu.RLock()
			k := s.k
			s.mu.RUnlock()
			anim, err := k.ChangePet(input.Name, input.Animation)
			if err != nil {
				return nil, nil, err
			}
			s.mu.Lock()
			s.registerTools(input.Name)
			s.mu.Unlock()
			s.v.Show(input.Name, anim)
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Switched to " + input.Name}}}, nil, nil
		})
	}
}

func (s *server) onConfigChange(cfg config.Config) {
	k := pet.NewKeeper(cfg)
	petName := k.DefaultPet()
	anim, _ := k.DefaultAnim(petName)
	s.mu.Lock()
	s.k = k
	s.registerTools(petName)
	s.mu.Unlock()
	s.v.Show(petName, anim)
}

func (s *server) Run(ctx context.Context) {
	defer s.cancel()

	mux := http.NewServeMux()
	mcpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return s.mcpServer }, nil)
	mux.Handle("/", mcpHandler)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

outer:
	for {
		ln, err := s.availablePort()
		if err != nil {
			s.logger.Error("failed to listen", "err", err)
			return
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

		for {
			select {
			case <-ctx.Done():
				httpSrv.Shutdown(context.Background())
				<-srvDone
				return
			case <-sigCh:
				s.cancel()
				httpSrv.Shutdown(context.Background())
				<-srvDone
				return
			case cfg, ok := <-s.cfgCh:
				if !ok {
					httpSrv.Shutdown(context.Background())
					<-srvDone
					return
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
					continue outer
				}
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
