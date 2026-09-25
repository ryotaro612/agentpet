package main

import (
	"context"
	"log"
	"os"

	"github.com/ryotaro612/agentpet/internal"
	"github.com/ryotaro612/agentpet/internal/config"
	"github.com/ryotaro612/agentpet/internal/server"
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

	cfgCh := make(chan config.Config)
	if _, err := config.NewConfigWatcher(ctx, args.ConfigFile, cfgCh, logger); err != nil {
		logger.Warn("failed to start config watcher", "err", err)
	}

	v := view.New(logger)
	s := server.NewServer(cfgCh, v, cfg.Port, cfg, logger, cancel)
	go s.Run(ctx)
	v.Run(ctx)
}
