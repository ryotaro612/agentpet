package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ryotaro612/agentpet/internal"
)

func main() {
	a, err := parse(os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}
	if err != nil {
		os.Exit(2)
	}

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
		fmt.Fprintf(os.Stderr, "error: cannot connect to %s: %v\n", endpoint, err)
		os.Exit(1)
	}
	defer session.Close()

	var runErr error
	switch a.command {
	case subcmdAnim:
		runErr = cmdAnimation(ctx, session, a.animation.animation, logger)
	case subcmdPet:
		runErr = cmdPet(ctx, session, a.pet.pet, a.pet.animation, logger)
	case subcmdPets:
		runErr = cmdPets(ctx, session, logger)
	case subcmdShow:
		runErr = cmdWindow(ctx, session, "show_window", logger)
	case subcmdHide:
		runErr = cmdWindow(ctx, session, "hide_window", logger)
	}
	if runErr != nil {
		fmt.Fprintln(os.Stderr, "error:", runErr)
		os.Exit(1)
	}
}

func cmdAnimation(ctx context.Context, session *mcp.ClientSession, name string, logger *slog.Logger) error {
	var toolName string
	if name != "" {
		toolName = "play_" + name
	} else {
		tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		if err != nil {
			return fmt.Errorf("listing tools: %w", err)
		}
		var animTools []string
		for _, t := range tools.Tools {
			if strings.HasPrefix(t.Name, "play_") {
				animTools = append(animTools, t.Name)
			}
		}
		if len(animTools) == 0 {
			return fmt.Errorf("no animation tools found")
		}
		candidates := animTools
		if current := currentAnimToolName(ctx, session); current != "" {
			others := make([]string, 0, len(animTools)-1)
			for _, t := range animTools {
				if t != current {
					others = append(others, t)
				}
			}
			if len(others) > 0 {
				candidates = others
			}
		}
		toolName = candidates[rand.IntN(len(candidates))]
	}
	logger.Debug("calling tool", "name", toolName)
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: toolName})
	if err != nil {
		return fmt.Errorf("calling %s: %w", toolName, err)
	}
	printResult(result)
	return nil
}

// currentAnimToolName returns the "play_<name>" tool name of the currently
// active animation by calling list_pets. Returns "" on any error so the caller
// can fall back gracefully.
func currentAnimToolName(ctx context.Context, session *mcp.ClientSession) string {
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_pets"})
	if err != nil || len(result.Content) == 0 {
		return ""
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return ""
	}
	var resp petsResponse
	if json.Unmarshal([]byte(text.Text), &resp) != nil {
		return ""
	}
	if resp.Active.Animation == "" {
		return ""
	}
	return "play_" + resp.Active.Animation
}

func cmdPet(ctx context.Context, session *mcp.ClientSession, pet, animation string, logger *slog.Logger) error {
	callArgs := map[string]any{"name": pet}
	if animation != "" {
		callArgs["animation"] = animation
	}
	logger.Debug("calling change_pet", "args", callArgs)
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "change_pet",
		Arguments: callArgs,
	})
	if err != nil {
		return fmt.Errorf("calling change_pet: %w", err)
	}
	printResult(result)
	return nil
}

type petsResponse struct {
	Active struct {
		Pet       string `json:"pet"`
		Animation string `json:"animation"`
	} `json:"active"`
	Pets []struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Animations  []struct {
			Name        string  `json:"name"`
			Description *string `json:"description"`
		} `json:"animations"`
	} `json:"pets"`
}

func cmdPets(ctx context.Context, session *mcp.ClientSession, logger *slog.Logger) error {
	logger.Debug("calling list_pets")
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_pets"})
	if err != nil {
		return fmt.Errorf("calling list_pets: %w", err)
	}
	if len(result.Content) == 0 {
		return nil
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return fmt.Errorf("unexpected content type")
	}
	var resp petsResponse
	if err := json.Unmarshal([]byte(text.Text), &resp); err != nil {
		fmt.Println(text.Text)
		return nil
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
