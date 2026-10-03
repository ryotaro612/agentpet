package server

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal"
	"github.com/ryotaro612/agentpet/internal/config"
)

func vetTool(cfgPath string) mcpTool {
	return mcpTool{
		tool: mcp.Tool{
			Name:        internal.ToolVet,
			Description: "Validate the config file and report any errors",
		},
		handler: func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			resp := internal.VetResponse{Valid: true}
			if _, err := config.LoadConfig(cfgPath); err != nil {
				resp.Valid = false
				resp.Description = err.Error()
			}
			b, err := json.Marshal(resp)
			if err != nil {
				return nil, nil, err
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
		},
	}
}
