package main

import (
	"log"
	"os"

	internal "github.com/ryotaro612/agentpet/internal"
	"github.com/webview/webview"
)

func main() {
	args, err := internal.ParseArgs()
	if err != nil {
		log.Fatal(err)
	}
	logger := internal.NewLogger(args.LogLevel())

	if args.ConfigFile != "" {
		if _, err := internal.LoadConfig(args.ConfigFile); err != nil {
			logger.Error("failed to load config", "err", err)
			os.Exit(1)
		}
	}
	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("Pet MCP")
	w.SetSize(480, 320, webview.HintNone)
	w.SetHtml("<h1>Hello, World!</h1>")
	w.Run()
}
