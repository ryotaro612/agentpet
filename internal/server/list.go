package server

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal"
	"github.com/ryotaro612/agentpet/internal/pet"
	"github.com/ryotaro612/agentpet/internal/view"
)

func listPetsTool(k pet.Keeper, v *view.View) mcpTool {
	return mcpTool{
		tool: mcp.Tool{
			Name:        internal.ToolListPets,
			Description: "List all available pets and their animations, indicating which are currently active",
		},
		handler: func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			allPets := k.AllPets()
			state := v.Current()

			pets := make([]internal.PetEntry, 0, len(allPets))
			for _, p := range allPets {
				anims := make([]internal.AnimEntry, 0, len(p.Animations))
				for _, a := range p.Animations {
					var desc *string
					if a.Description != "" {
						desc = &a.Description
					}
					anims = append(anims, internal.AnimEntry{Name: a.Name, Description: desc})
				}
				pets = append(pets, internal.PetEntry{Name: p.Name, Animations: anims})
			}
			b, err := json.Marshal(internal.PetsResponse{
				Active: internal.ActiveEntry{Pet: state.Pet, Animation: state.Animation},
				Pets:   pets,
			})
			if err != nil {
				return nil, nil, err
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
		},
	}
}
