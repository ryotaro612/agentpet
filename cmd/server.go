package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal"
)

func runWithSession(a parsedArgs) error {
	logLevel := slog.LevelInfo
	if a.verbose {
		logLevel = slog.LevelDebug
	}
	logger := internal.NewLogger(logLevel)

	endpoint := fmt.Sprintf("http://localhost:%d/", a.port)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             endpoint,
		DisableStandaloneSSE: true,
	}

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: appName, Version: "v0.0.1"}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("cannot connect to %s: %w", endpoint, err)
	}
	defer session.Close()

	switch a.command {
	case subcmdAnim:
		return cmdAnimation(ctx, session, a.animation.animation, logger)
	case subcmdPet:
		return cmdPet(ctx, session, a.pet.pet, a.pet.animation, logger)
	case subcmdPets:
		return cmdPets(ctx, session, logger)
	case subcmdShow:
		return cmdWindow(ctx, session, internal.ToolShowWindow, logger)
	case subcmdHide:
		return cmdWindow(ctx, session, internal.ToolHideWindow, logger)
	}
	return nil
}

func cmdAnimation(ctx context.Context, session *mcp.ClientSession, name string, logger *slog.Logger) error {
	toolName := internal.ToolShuffle
	if name != "" {
		toolName = internal.ToolPlayPrefix + name
	}
	logger.Debug("calling tool", "name", toolName)
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: toolName})
	if err != nil {
		return fmt.Errorf("calling %s: %w", toolName, err)
	}
	printResult(result)
	return nil
}

func cmdPet(ctx context.Context, session *mcp.ClientSession, pet, animation string, logger *slog.Logger) error {
	if pet == "" {
		resp, err := listPets(ctx, session)
		if err != nil {
			return err
		}
		others := make([]string, 0, len(resp.Pets))
		for _, p := range resp.Pets {
			if p.Name != resp.Active.Pet {
				others = append(others, p.Name)
			}
		}
		if len(others) == 0 {
			return fmt.Errorf("no other pets to switch to")
		}
		pet = others[rand.IntN(len(others))]
	}
	input := internal.ChangePetInput{Name: pet, Animation: animation}
	callArgs := map[string]any{"name": input.Name}
	if input.Animation != "" {
		callArgs["animation"] = input.Animation
	}
	logger.Debug("calling change_pet", "args", callArgs)
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      internal.ToolChangePet,
		Arguments: callArgs,
	})
	if err != nil {
		return fmt.Errorf("calling change_pet: %w", err)
	}
	printResult(result)
	return nil
}

func listPets(ctx context.Context, session *mcp.ClientSession) (internal.PetsResponse, error) {
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: internal.ToolListPets})
	if err != nil {
		return internal.PetsResponse{}, fmt.Errorf("listing pets: %w", err)
	}
	if len(result.Content) == 0 {
		return internal.PetsResponse{}, fmt.Errorf("empty response from list_pets")
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return internal.PetsResponse{}, fmt.Errorf("unexpected content type from list_pets")
	}
	var resp internal.PetsResponse
	if err := json.Unmarshal([]byte(text.Text), &resp); err != nil {
		return internal.PetsResponse{}, fmt.Errorf("parsing pets response: %w", err)
	}
	return resp, nil
}

func cmdPets(ctx context.Context, session *mcp.ClientSession, logger *slog.Logger) error {
	logger.Debug("calling list_pets")
	resp, err := listPets(ctx, session)
	if err != nil {
		return err
	}
	for _, p := range resp.Pets {
		active := ""
		if p.Name == resp.Active.Pet {
			active = " (active)"
		}
		fmt.Printf("%s%s\n", p.Name, active)
		for _, a := range p.Animations {
			animActive := ""
			if p.Name == resp.Active.Pet && a.Name == resp.Active.Animation {
				animActive = " *"
			}
			desc := ""
			if a.Description != nil {
				desc = "  — " + *a.Description
			}
			fmt.Printf("  %s%s%s\n", a.Name, animActive, desc)
		}
	}
	return nil
}

func cmdWindow(ctx context.Context, session *mcp.ClientSession, toolName string, logger *slog.Logger) error {
	logger.Debug("calling tool", "name", toolName)
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: toolName})
	if err != nil {
		return fmt.Errorf("calling %s: %w", toolName, err)
	}
	printResult(result)
	return nil
}

func printResult(result *mcp.CallToolResult) {
	for _, c := range result.Content {
		if t, ok := c.(*mcp.TextContent); ok {
			fmt.Println(t.Text)
		}
	}
}
