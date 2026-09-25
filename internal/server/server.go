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
	// start check
	restartPort chan int // receives a new port to restart the HTTP server on
	// end check
}

func NewServer(cfgCh <-chan config.Config, v *view.View, port int, cfg config.Config, logger *slog.Logger, cancel context.CancelFunc) *server {
	s := &server{
		cfgCh: cfgCh,
		v:     v,
		port:  port,
		mcpServer: mcp.NewServer(&mcp.Implementation{Name: "agentpet", Version: "v0.0.1"}, &mcp.ServerOptions{
			Capabilities: &mcp.ServerCapabilities{
				Tools: &mcp.ToolCapabilities{ListChanged: true},
			},
		}),
		logger:      logger,
		cancel:      cancel,
		k:           pet.NewKeeper(cfg),
		restartPort: make(chan int, 1),
	}

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "show_window",
		Description: "Show the pet window",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		if s.v != nil {
			s.v.ShowWindow()
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Window is now visible"}}}, nil, nil
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "hide_window",
		Description: "Hide the pet window",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		if s.v != nil {
			s.v.HideWindow()
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Window is now hidden"}}}, nil, nil
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_pets",
		Description: "List all available pets and their animations, indicating which are currently active",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		s.mu.RLock()
		allPets := s.k.AllPets()
		currentPet := s.k.CurrentPetName()
		currentAnim := s.k.CurrentAnimation().Name
		s.mu.RUnlock()

		type animEntry struct {
			Name        string `json:"name"`
			Description string `json:"description,omitempty"`
			Active      bool   `json:"active,omitempty"`
		}
		type petEntry struct {
			Name       string      `json:"name"`
			Active     bool        `json:"active,omitempty"`
			Animations []animEntry `json:"animations"`
		}
		type response struct {
			Pets []petEntry `json:"pets"`
		}

		pets := make([]petEntry, 0, len(allPets))
		for _, p := range allPets {
			anims := make([]animEntry, 0, len(p.Animations))
			for _, a := range p.Animations {
				anims = append(anims, animEntry{
					Name:        a.Name,
					Description: a.Description,
					Active:      p.Name == currentPet && a.Name == currentAnim,
				})
			}
			pets = append(pets, petEntry{
				Name:       p.Name,
				Active:     p.Name == currentPet,
				Animations: anims,
			})
		}
		b, err := json.Marshal(response{Pets: pets})
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
	})

	s.registerTools()
	return s
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

func (s *server) animToolNames() []string {
	anims := s.k.Animations()
	names := make([]string, len(anims))
	for i, a := range anims {
		names[i] = "play_" + a.Name
	}
	return names
}

// registerTools must be called with s.mu held for writing.
func (s *server) registerTools() {
	s.mcpServer.RemoveTools(append(s.animToolNames(), "change_pet")...)

	for _, a := range s.k.Animations() {
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
			s.mu.Lock()
			newK, err := s.k.PlayAnimation(animName)
			if err == nil {
				s.k = newK
			}
			anim := s.k.CurrentAnimation()
			s.mu.Unlock()
			if err != nil {
				return nil, nil, err
			}
			if s.v != nil {
				s.v.Show(anim)
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Playing " + animName}}}, nil, nil
		})
	}

	if otherPets := s.k.OtherPetNames(); len(otherPets) > 0 {
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        "change_pet",
			Description: "Switch the active pet",
			InputSchema: buildChangePetSchema(otherPets),
		}, func(_ context.Context, _ *mcp.CallToolRequest, input changePetInput) (*mcp.CallToolResult, any, error) {
			s.mu.Lock()
			newK, err := s.k.ChangePet(input.Name, input.Animation)
			if err != nil {
				s.mu.Unlock()
				return nil, nil, err
			}
			s.k = newK
			s.registerTools()
			anim := s.k.CurrentAnimation()
			s.mu.Unlock()
			if s.v != nil {
				s.v.Show(anim)
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Switched to " + input.Name}}}, nil, nil
		})
	}
}

func (s *server) onConfigChange(cfg config.Config) {
	s.mu.Lock()
	s.k = pet.NewKeeper(cfg)
	s.registerTools()
	anim := s.k.CurrentAnimation()
	active := s.port
	s.mu.Unlock()
	if s.v != nil {
		s.v.Show(anim)
	}

	if cfg.Port != 0 && cfg.Port != active {
		select {
		case s.restartPort <- cfg.Port:
		default:
		}
	}
}

func (s *server) Run(ctx context.Context) {
	defer s.cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case cfg, ok := <-s.cfgCh:
				if !ok {
					return
				}
				s.onConfigChange(cfg)
			}
		}
	}()

	port, err := s.availablePort()
	if err != nil {
		s.logger.Error("failed to find available port", "err", err)
		return
	}

	mux := http.NewServeMux()
	mcpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return s.mcpServer }, nil)
	mux.Handle("/", mcpHandler)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	if s.v != nil {
		s.mu.RLock()
		anim := s.k.CurrentAnimation()
		s.mu.RUnlock()
		s.v.Show(anim)
	}

	for {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			s.logger.Error("failed to listen", "port", port, "err", err)
			return
		}

		s.mu.Lock()
		s.port = ln.Addr().(*net.TCPAddr).Port
		s.mu.Unlock()

		s.logger.Info("MCP server listening", "addr", ln.Addr().String())

		httpSrv := &http.Server{Handler: mux}
		srvDone := make(chan struct{})
		go func() {
			defer close(srvDone)
			if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
				s.logger.Error("server error", "err", err)
			}
		}()

		select {
		case <-ctx.Done():
		case <-sigCh:
			s.cancel()
		case newPort := <-s.restartPort:
			httpSrv.Shutdown(context.Background())
			<-srvDone
			port = newPort
			s.logger.Info("MCP server restarting", "port", newPort)
			continue
		}
		httpSrv.Shutdown(context.Background())
		<-srvDone
		return
	}
}

func (s *server) availablePort() (int, error) {
	if s.port != 0 {
		return s.port, nil
	}
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}
