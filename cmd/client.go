package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	fs := flag.NewFlagSet("agentpet-client", flag.ExitOnError)
	port := fs.Int("p", 0, "MCP server port")
	verbose := fs.Bool("v", false, "verbose output")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: agentpet-client -p <port> [-v] <subcommand> [args]\n\n")
		fmt.Fprintf(os.Stderr, "Subcommands:\n")
		fmt.Fprintf(os.Stderr, "  animation [name]       change to the named animation, or a random one if omitted\n")
		fmt.Fprintf(os.Stderr, "  pet <name> [animation] switch to a different pet, optionally activating an animation\n")
		fmt.Fprintf(os.Stderr, "  pets                   list available pets and their animations\n")
		fmt.Fprintf(os.Stderr, "  show                   show the pet window\n")
		fmt.Fprintf(os.Stderr, "  hide                   hide the pet window\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if *port == 0 {
		fmt.Fprintln(os.Stderr, "error: -p <port> is required")
		fs.Usage()
		os.Exit(2)
	}

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		os.Exit(2)
	}

	endpoint := fmt.Sprintf("http://localhost:%d/", *port)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             endpoint,
		DisableStandaloneSSE: true,
	}

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "agentpet-client", Version: "v0.0.1"}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot connect to %s: %v\n", endpoint, err)
		os.Exit(1)
	}
	defer session.Close()

	var runErr error
	switch args[0] {
	case "animation":
		runErr = cmdAnimation(ctx, session, args[1:], *verbose)
	case "pet":
		runErr = cmdPet(ctx, session, args[1:], *verbose)
	case "pets":
		runErr = cmdPets(ctx, session, *verbose)
	case "show":
		runErr = cmdWindow(ctx, session, "show_window", *verbose)
	case "hide":
		runErr = cmdWindow(ctx, session, "hide_window", *verbose)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown subcommand %q\n", args[0])
		fs.Usage()
		os.Exit(2)
	}
	if runErr != nil {
		fmt.Fprintln(os.Stderr, "error:", runErr)
		os.Exit(1)
	}
}

func cmdAnimation(ctx context.Context, session *mcp.ClientSession, args []string, verbose bool) error {
	var toolName string
	if len(args) > 0 {
		toolName = "anim_" + args[0]
	} else {
		tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		if err != nil {
			return fmt.Errorf("listing tools: %w", err)
		}
		var animTools []string
		for _, t := range tools.Tools {
			if strings.HasPrefix(t.Name, "anim_") {
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
	if verbose {
		fmt.Println("calling tool:", toolName)
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: toolName})
	if err != nil {
		return fmt.Errorf("calling %s: %w", toolName, err)
	}
	printResult(result)
	return nil
}

// currentAnimToolName returns the "anim_<name>" tool name of the currently
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
	for _, p := range resp.Pets {
		if !p.Active {
			continue
		}
		for _, a := range p.Animations {
			if a.Active {
				return "anim_" + a.Name
			}
		}
	}
	return ""
}

func cmdPet(ctx context.Context, session *mcp.ClientSession, args []string, verbose bool) error {
	if len(args) == 0 {
		return fmt.Errorf("pet: <name> is required")
	}
	petName := args[0]
	callArgs := map[string]any{"name": petName}
	if len(args) > 1 {
		callArgs["animation"] = args[1]
	}
	if verbose {
		fmt.Printf("calling change_pet with %v\n", callArgs)
	}
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
	Pets []struct {
		Name       string `json:"name"`
		Active     bool   `json:"active"`
		Animations []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Active      bool   `json:"active"`
		} `json:"animations"`
	} `json:"pets"`
}

func cmdPets(ctx context.Context, session *mcp.ClientSession, verbose bool) error {
	if verbose {
		fmt.Println("calling list_pets")
	}
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
		if p.Active {
			active = " (active)"
		}
		fmt.Printf("%s%s\n", p.Name, active)
		for _, a := range p.Animations {
			animActive := ""
			if a.Active {
				animActive = " *"
			}
			desc := ""
			if a.Description != "" {
				desc = "  — " + a.Description
			}
			fmt.Printf("  %s%s%s\n", a.Name, animActive, desc)
		}
	}
	return nil
}

func cmdWindow(ctx context.Context, session *mcp.ClientSession, toolName string, verbose bool) error {
	if verbose {
		fmt.Println("calling tool:", toolName)
	}
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
