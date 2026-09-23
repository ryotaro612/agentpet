package main

import (
	"context"
	"log"
	"os"

	internal "github.com/ryotaro612/agentpet/internal"
	"github.com/ryotaro612/agentpet/internal/config"
	"github.com/ryotaro612/agentpet/internal/view"
)

func main() {
	args, err := internal.ParseArgs()
	if err != nil {
		log.Fatal(err)
	}
	logger := internal.NewLogger(args.LogLevel())

	cfg, err := config.LoadConfig(args.ConfigFile)
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher := config.NewConfigWatcher(ctx, args.ConfigFile, cfg, logger)

	v := view.New(logger)
	s := internal.NewServer(watcher, v, cfg.Port, logger, cancel)
	go s.Run(ctx)
	v.Run(ctx)
}
