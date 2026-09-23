package internal

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
	"github.com/ryotaro612/agentpet/internal/view"
)

type nameInput struct {
	Name string `json:"name"`
}

type server struct {
	watcher   *config.Watcher
	v         *view.View
	portNum   int
	mcpServer *mcp.Server
	logger    *slog.Logger
	cancel    context.CancelFunc
	toolsMu   sync.Mutex
	k         *keeper
}

func NewServer(w *config.Watcher, v *view.View, port int, logger *slog.Logger, cancel context.CancelFunc) *server {
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
		k:      newKeeper(w.Get()),
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

	s.registerTools()
	return s
}

func toViewAnim(a animation) view.Anim {
	fps, fpsErr := a.calcFps()
	dim := a.windowDim()
	filePath, _ := a.absFilePath()
	return view.Anim{
		Name:     a.name,
		FPS:      fps,
		FPSErr:   fpsErr,
		FrameW:   a.frame.width,
		FrameH:   a.frame.height,
		WinW:     dim.width,
		WinH:     dim.height,
		FilePath: filePath,
	}
}

func buildEnumSchema(param, description string, values []string) json.RawMessage {
	type property struct {
		Type        string   `json:"type"`
		Description string   `json:"description"`
		Enum        []string `json:"enum"`
	}
	type schema struct {
		Type       string              `json:"type"`
		Properties map[string]property `json:"properties"`
		Required   []string            `json:"required"`
	}
	b, _ := json.Marshal(schema{
		Type:       "object",
		Properties: map[string]property{param: {Type: "string", Description: description, Enum: values}},
		Required:   []string{param},
	})
	return json.RawMessage(b)
}

func (s *server) registerTools() {
	s.toolsMu.Lock()
	defer s.toolsMu.Unlock()

	animNames := s.k.animationNames()
	otherPets := s.k.otherPetNames()

	s.mcpServer.RemoveTools("play_animation", "change_pet")

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "play_animation",
		Description: "Play the named animation for the active pet",
		InputSchema: buildEnumSchema("name", "Name of the animation to play", animNames),
	}, func(_ context.Context, _ *mcp.CallToolRequest, input nameInput) (*mcp.CallToolResult, any, error) {
		anim, err := s.k.playAnimation(input.Name)
		if err != nil {
			return nil, nil, err
		}
		if s.v != nil && anim != nil {
			s.v.Show(toViewAnim(*anim))
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Playing " + input.Name}}}, nil, nil
	})

	if len(otherPets) > 0 {
		mcp.AddTool(s.mcpServer, &mcp.Tool{
			Name:        "change_pet",
			Description: "Switch the active pet",
			InputSchema: buildEnumSchema("name", "Name of the pet to switch to", otherPets),
		}, func(_ context.Context, _ *mcp.CallToolRequest, input nameInput) (*mcp.CallToolResult, any, error) {
			anim, err := s.k.changePet(input.Name)
			if err != nil {
				return nil, nil, err
			}
			s.registerTools()
			if s.v != nil && anim != nil {
				s.v.Show(toViewAnim(*anim))
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Switched to " + input.Name}}}, nil, nil
		})
	}
}

func (s *server) onConfigChange(cfg config.Config) {
	s.k.rebuildFrom(cfg)
	s.registerTools()
	if s.v != nil {
		if a := s.k.currentAnimation(); a != nil {
			s.v.Show(toViewAnim(*a))
		}
	}
}

func (s *server) serveImage(w http.ResponseWriter, r *http.Request) {
	anim := s.k.currentAnimation()
	if anim == nil {
		http.NotFound(w, r)
		return
	}
	path, err := anim.absFilePath()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, path)
}

func (s *server) Run(ctx context.Context) {
	defer s.cancel()

	if err := s.watcher.Watch(s.onConfigChange); err != nil {
		s.logger.Warn("failed to start config watcher", "err", err)
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
	addr := ln.Addr().String()
	s.logger.Info("MCP server listening", "addr", addr)

	if s.v != nil {
		if a := s.k.currentAnimation(); a != nil {
			s.v.Show(toViewAnim(*a))
		}
	}

	mux := http.NewServeMux()
	mcpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return s.mcpServer }, nil)
	mux.HandleFunc("/image", s.serveImage)
	mux.Handle("/", mcpHandler)
	httpSrv := &http.Server{Handler: mux}

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
