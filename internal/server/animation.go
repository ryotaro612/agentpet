package server

import (
	"context"
	"math/rand/v2"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal"
	"github.com/ryotaro612/agentpet/internal/pet"
	"github.com/ryotaro612/agentpet/internal/view"
)

func shuffleAnimationTool(s *server) mcpTool {
	return mcpTool{
		tool: mcp.Tool{
			Name:        internal.ToolShuffle,
			Description: "Play a randomly selected animation, different from the current one",
		},
		handler: func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			currentPet := s.v.CurrentPet()
			s.mu.RLock()
			k := s.k
			s.mu.RUnlock()
			currentAnim := s.v.CurrentAnimation()
			anims := k.Animations(currentPet)
			candidates := make([]string, 0, len(anims))
			for _, a := range anims {
				if a.Name != currentAnim {
					candidates = append(candidates, a.Name)
				}
			}
			if len(candidates) == 0 {
				candidates = []string{currentAnim}
			}
			return playAndShow(s.v, k, currentPet, candidates[rand.IntN(len(candidates))])
		},
	}
}

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
				currentPet := s.v.CurrentPet()
				s.mu.RLock()
				k := s.k
				s.mu.RUnlock()
				return playAndShow(s.v, k, currentPet, animName)
			},
		})
	}
	return tools
}

func playAndShow(v *view.View, k pet.Keeper, petName, animName string) (*mcp.CallToolResult, any, error) {
	anim, err := k.PlayAnimation(petName, animName)
	if err != nil {
		return nil, nil, err
	}
	v.Show(petName, anim)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Playing " + animName}}}, nil, nil
}
