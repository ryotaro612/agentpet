package main

import (
	"flag"
	"fmt"
	"io"
)

type args struct {
	Port    int
	Verbose bool
	Command string
	Rest    []string
}

func parse(argv []string, output io.Writer) (args, error) {
	fs := flag.NewFlagSet("agentpet-client", flag.ContinueOnError)
	fs.SetOutput(output)
	port := fs.Int("p", 0, "MCP server `port`")
	verbose := fs.Bool("v", false, "verbose output")

	animationCmd := flag.NewFlagSet("animation", flag.ContinueOnError)
	animationCmd.SetOutput(output)
	petCmd := flag.NewFlagSet("pet", flag.ContinueOnError)
	petCmd.SetOutput(output)
	petsCmd := flag.NewFlagSet("pets", flag.ContinueOnError)
	petsCmd.SetOutput(output)
	showCmd := flag.NewFlagSet("show", flag.ContinueOnError)
	showCmd.SetOutput(output)
	hideCmd := flag.NewFlagSet("hide", flag.ContinueOnError)
	hideCmd.SetOutput(output)

	if err := fs.Parse(argv); err != nil {
		return args{}, err
	}
	if *port == 0 {
		fmt.Fprintln(output, "error: -p <port> is required")
		fs.Usage()
		return args{}, fmt.Errorf("-p <port> is required")
	}

	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprintln(output, "error: subcommand is required")
		fs.Usage()
		return args{}, fmt.Errorf("subcommand is required")
	}

	cmd := rest[0]
	var sub *flag.FlagSet
	switch cmd {
	case "animation":
		sub = animationCmd
	case "pet":
		sub = petCmd
	case "pets":
		sub = petsCmd
	case "show":
		sub = showCmd
	case "hide":
		sub = hideCmd
	default:
		fmt.Fprintf(output, "error: unknown subcommand %q\n", cmd)
		return args{}, fmt.Errorf("unknown subcommand %q", cmd)
	}

	if err := sub.Parse(rest[1:]); err != nil {
		return args{}, err
	}
	return args{Port: *port, Verbose: *verbose, Command: cmd, Rest: sub.Args()}, nil
}
