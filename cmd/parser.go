package main

import (
	"flag"
	"fmt"
	"io"
)

const (
	appName          = "petowner"
	subcmdAnim       = "animation"
	subcmdCompletion = "completion"
	subcmdPet        = "pet"
	subcmdPets       = "pets"
	subcmdShow       = "show"
	subcmdHide       = "hide"
	subcmdVet        = "vet"
)

type animationArgs struct {
	animation string
}

type completionArgs struct {
	shell string
}

type petArgs struct {
	verbose bool
	pet     string
	animationArgs
}

type parsedArgs struct {
	port       int
	command    string
	verbose    bool
	animation  animationArgs
	completion completionArgs
	pet        petArgs
}

func parse(argv []string, output io.Writer) (parsedArgs, error) {
	newFlagSet := func(name string) *flag.FlagSet {
		fs := flag.NewFlagSet(name, flag.ContinueOnError)
		fs.SetOutput(output)
		return fs
	}

	fs := newFlagSet(appName)
	port := fs.Int("p", 0, "MCP server `port`")
	verbose := fs.Bool("v", false, "verbose output")

	animationCmd := newFlagSet(subcmdAnim)
	completionCmd := newFlagSet(subcmdCompletion)
	petCmd := newFlagSet(subcmdPet)
	petsCmd := newFlagSet(subcmdPets)
	showCmd := newFlagSet(subcmdShow)
	hideCmd := newFlagSet(subcmdHide)
	vetCmd := newFlagSet(subcmdVet)

	fs.Usage = func() {
		fmt.Fprintf(output, "Usage: %s -p <port> [-v] <subcommand> [args]\n\n", appName)
		fmt.Fprintf(output, "Flags:\n")
		fs.PrintDefaults()
		fmt.Fprintf(output, "\nSubcommands:\n")
		fmt.Fprintf(output, "  %s [name]\n", subcmdAnim)
		fmt.Fprintf(output, "  %s zsh\n", subcmdCompletion)
		fmt.Fprintf(output, "  %s <name> [animation]\n", subcmdPet)
		fmt.Fprintf(output, "  %s\n", subcmdPets)
		fmt.Fprintf(output, "  %s\n", subcmdShow)
		fmt.Fprintf(output, "  %s\n", subcmdHide)
		fmt.Fprintf(output, "  %s\n", subcmdVet)
	}

	if err := fs.Parse(argv); err != nil {
		return parsedArgs{}, err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprintln(output, "error: subcommand is required")
		fs.Usage()
		return parsedArgs{}, fmt.Errorf("subcommand is required")
	}

	cmd := rest[0]
	subArgs := rest[1:]
	result := parsedArgs{port: *port, command: cmd, verbose: *verbose}
	if cmd != subcmdCompletion && *port == 0 {
		fmt.Fprintln(output, "error: -p <port> is required")
		fs.Usage()
		return parsedArgs{}, fmt.Errorf("-p <port> is required")
	}

	switch cmd {
	case subcmdAnim:
		if err := animationCmd.Parse(subArgs); err != nil {
			return parsedArgs{}, err
		}
		name := ""
		if len(animationCmd.Args()) > 0 {
			name = animationCmd.Args()[0]
		}
		result.animation = animationArgs{animation: name}
	case subcmdCompletion:
		if err := completionCmd.Parse(subArgs); err != nil {
			return parsedArgs{}, err
		}
		if len(completionCmd.Args()) != 1 || completionCmd.Args()[0] != "zsh" {
			return parsedArgs{}, fmt.Errorf("usage: %s %s zsh", appName, subcmdCompletion)
		}
		result.completion = completionArgs{shell: completionCmd.Args()[0]}
	case subcmdPet:
		if err := petCmd.Parse(subArgs); err != nil {
			return parsedArgs{}, err
		}
		if len(petCmd.Args()) > 0 {
			result.pet.pet = petCmd.Args()[0]
		}
		if len(petCmd.Args()) > 1 {
			result.pet.animation = petCmd.Args()[1]
		}
		result.pet.verbose = *verbose
	case subcmdPets:
		if err := petsCmd.Parse(subArgs); err != nil {
			return parsedArgs{}, err
		}
	case subcmdShow:
		if err := showCmd.Parse(subArgs); err != nil {
			return parsedArgs{}, err
		}
	case subcmdHide:
		if err := hideCmd.Parse(subArgs); err != nil {
			return parsedArgs{}, err
		}
	case subcmdVet:
		if err := vetCmd.Parse(subArgs); err != nil {
			return parsedArgs{}, err
		}
	default:
		fmt.Fprintf(output, "error: unknown subcommand %q\n", cmd)
		return parsedArgs{}, fmt.Errorf("unknown subcommand %q", cmd)
	}

	return result, nil
}
