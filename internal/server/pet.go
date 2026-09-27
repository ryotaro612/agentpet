package server

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal"
)

type changePetMcpTool struct {
	tool    mcp.Tool
	handler mcp.ToolHandlerFor[internal.ChangePetInput, any]
}

func changePetTool(petName string, s *server) (changePetMcpTool, bool) {
	otherPets := s.k.OtherPetNames(petName)
	if len(otherPets) == 0 {
		return changePetMcpTool{}, false
	}
	schema := buildChangePetSchema(otherPets)
	return changePetMcpTool{
		tool: mcp.Tool{
			Name:        internal.ToolChangePet,
			Description: "Switch the active pet",
			InputSchema: schema,
		},
		handler: func(_ context.Context, _ *mcp.CallToolRequest, input internal.ChangePetInput) (*mcp.CallToolResult, any, error) {
			s.mu.RLock()
			k := s.k
			s.mu.RUnlock()
			anim, err := k.ChangePet(input.Name, input.Animation)
			if err != nil {
				return nil, nil, err
			}
			s.mu.Lock()
			currentPet := s.v.Current().Pet
			oldAnimTools := playAnimationTools(currentPet, s).names()
			s.registerTools(input.Name, oldAnimTools)
			s.mu.Unlock()
			s.v.Show(input.Name, anim)
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Switched to " + input.Name}}}, nil, nil
		},
	}, true
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
