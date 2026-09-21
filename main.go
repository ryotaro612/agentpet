package main

import "github.com/webview/webview"
import internal "github.com/ryotaro612/agentpet/internal"

func main() {
	internal.ParseCmd()
	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("Pet MCP")
	w.SetSize(480, 320, webview.HintNone)
	w.SetHtml("<h1>Hello, World!</h1>")
	w.Run()
}
