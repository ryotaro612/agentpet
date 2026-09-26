package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal/view"
)

type windowTool struct {
	tool    mcp.Tool
	handler mcp.ToolHandlerFor[struct{}, any]
}

func showWindowTool(v *view.View) windowTool {
	return newWindowTool("show_window", "Show the pet window", v.ShowWindow, "Window is now visible")
}

func hideWindowTool(v *view.View) windowTool {
	return newWindowTool("hide_window", "Hide the pet window", v.HideWindow, "Window is now hidden")
}

func newWindowTool(name, description string, action func(), result string) windowTool {
	return windowTool{
		tool: mcp.Tool{Name: name, Description: description},
		handler: func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			action()
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: result}}}, nil, nil
		},
	}
}
