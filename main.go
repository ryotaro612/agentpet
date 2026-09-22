package main

import (
	"context"
	"log"
	"os"

	internal "github.com/ryotaro612/agentpet/internal"
)

func main() {
	args, err := internal.ParseArgs()
	if err != nil {
		log.Fatal(err)
	}
	logger := internal.NewLogger(args.LogLevel())

	cfg, err := internal.LoadConfig(args.ConfigFile)
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	keeper, err := internal.NewKeeper(args.ConfigFile)
	if err != nil {
		logger.Error("failed to initialize keeper", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := internal.NewServer(keeper, internal.NewView(), cfg.Port(), logger, cancel)
	s.Run(ctx)
}
