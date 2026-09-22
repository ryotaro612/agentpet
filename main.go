package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()

	s := internal.NewServer(internal.NewKeeper(), internal.NeedView(), cfg.Server.Port, args.ConfigFile, logger)
	s.Run(ctx, cancel)
}
