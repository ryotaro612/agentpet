package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal"
)

func playAnimationTools(petName string, s *server) mcpTools {
	anims := s.k.Animations(petName)
	tools := make([]mcpTool, 0, len(anims))
	for _, a := range anims {
		name := internal.ToolPlayPrefix + a.Name
		desc := a.Description
		if desc == "" {
			desc = "Play the " + a.Name + " animation"
		}
		animName := a.Name
		tools = append(tools, mcpTool{
			tool: mcp.Tool{Name: name, Description: desc},
			handler: func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
				currentPet := s.v.Current().Pet
				s.mu.RLock()
				k := s.k
				s.mu.RUnlock()
				anim, err := k.PlayAnimation(currentPet, animName)
				if err != nil {
					return nil, nil, err
				}
				s.v.Show(currentPet, anim)
				return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Playing " + animName}}}, nil, nil
			},
		})
	}
	return tools
}
