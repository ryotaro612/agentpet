package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ryotaro612/agentpet/internal/config"
)

func main() {
	a, err := parse(os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}
	if err != nil {
		os.Exit(2)
	}
	var runErr error
	switch a.command {
	case subcmdCompletion, subcmdVet:
		runErr = runLocal(a)
	default:
		runErr = runWithSession(a)
	}
	if runErr != nil {
		fmt.Fprintln(os.Stderr, "error:", runErr)
		os.Exit(1)
	}
}

func runLocal(a parsedArgs) error {
	switch a.command {
	case subcmdCompletion:
		fmt.Print(zshCompletion)
	case subcmdVet:
		return cmdVet(a.vet.configFile)
	}
	return nil
}

func cmdVet(configFile string) error {
	if _, err := config.LoadConfig(configFile); err != nil {
		return err
	}
	fmt.Println("config is valid")
	return nil
}
