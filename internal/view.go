package internal

import (
	"github.com/webview/webview"
)

type view struct {
}

func RunWindow() {
	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("Pet MCP")
	w.SetSize(480, 320, webview.HintNone)
	win := w.Window()
	w.Bind("moveWindow", func(dx, dy float64) {
		MoveWindow(win, dx, dy)
	})
	w.SetHtml(`<!DOCTYPE html>
<html>
<head>
<style>
html, body {
  margin: 0;
  padding: 0;
  background: transparent;
  user-select: none;
  cursor: default;
}
</style>
</head>
<body>
<h1>Hello, World!</h1>
<script>
let dragging = false, lastX, lastY;
document.addEventListener('mousedown', e => {
  dragging = true; lastX = e.screenX; lastY = e.screenY;
});
document.addEventListener('mousemove', e => {
  if (!dragging) return;
  window.moveWindow(e.screenX - lastX, e.screenY - lastY);
  lastX = e.screenX; lastY = e.screenY;
});
document.addEventListener('mouseup', () => { dragging = false; });
</script>
</body>
</html>`)
	HideToolbar(w.Window())
	MakeTransparent(w.Window())
	w.Run()
}
