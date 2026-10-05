package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal"
	"github.com/ryotaro612/agentpet/internal/view"
)

type mcpTool struct {
	tool    mcp.Tool
	handler mcp.ToolHandlerFor[struct{}, any]
}

func (t mcpTool) toolName() string {
	return t.tool.Name
}

type mcpTools []mcpTool

func (ts mcpTools) names() []string {
	names := make([]string, len(ts))
	for i, t := range ts {
		names[i] = t.toolName()
	}
	return names
}

// registerTools must be called with s.mu held for writing.
func (s *server) registerTools(petName string, oldAnimTools []string) {
	lp := listPetsTool(s.k, s.v)
	s.mcpServer.RemoveTools(append(oldAnimTools, internal.ToolChangePet, lp.toolName())...)
	mcp.AddTool(s.mcpServer, &lp.tool, lp.handler)
	for _, t := range playAnimationTools(petName, s) {
		mcp.AddTool(s.mcpServer, &t.tool, t.handler)
	}
	if t, ok := changePetTool(petName, s); ok {
		mcp.AddTool(s.mcpServer, &t.tool, t.handler)
	}
}

func showWindowTool(v *view.View) mcpTool {
	return newWindowTool(internal.ToolShowWindow, "Show the pet window", v.ShowWindow, "Window is now visible")
}

func hideWindowTool(v *view.View) mcpTool {
	return newWindowTool(internal.ToolHideWindow, "Hide the pet window", v.HideWindow, "Window is now hidden")
}

func newWindowTool(name, description string, action func(), result string) mcpTool {
	return mcpTool{
		tool: mcp.Tool{Name: name, Description: description},
		handler: func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			action()
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil, nil
		},
	}
}
