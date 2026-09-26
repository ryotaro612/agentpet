package server

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal/pet"
	"github.com/ryotaro612/agentpet/internal/view"
)

func listPetsTool(k pet.Keeper, v *view.View) mcpTool {
	return mcpTool{
		tool: mcp.Tool{
			Name:        "list_pets",
			Description: "List all available pets and their animations, indicating which are currently active",
		},
		handler: func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
			allPets := k.AllPets()
			currentPet, currentAnim := v.Current()

			type activeEntry struct {
				Pet       string `json:"pet"`
				Animation string `json:"animation"`
			}
			type animEntry struct {
				Name        string  `json:"name"`
				Description *string `json:"description"`
			}
			type petEntry struct {
				Name        string      `json:"name"`
				Description *string     `json:"description"`
				Animations  []animEntry `json:"animations"`
			}
			type response struct {
				Active activeEntry `json:"active"`
				Pets   []petEntry  `json:"pets"`
			}

			pets := make([]petEntry, 0, len(allPets))
			for _, p := range allPets {
				anims := make([]animEntry, 0, len(p.Animations))
				for _, a := range p.Animations {
					var desc *string
					if a.Description != "" {
						desc = &a.Description
					}
					anims = append(anims, animEntry{Name: a.Name, Description: desc})
				}
				pets = append(pets, petEntry{Name: p.Name, Animations: anims})
			}
			b, err := json.Marshal(response{
				Active: activeEntry{Pet: currentPet, Animation: currentAnim},
				Pets:   pets,
			})
			if err != nil {
				return nil, nil, err
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
		},
	}
}
