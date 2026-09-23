package internal

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/webview/webview"
)

type view struct {
	w        webview.WebView
	animCh   chan animation
	filePort int
	readyCh  chan struct{}
}

func NewView() *view {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	go http.Serve(ln, http.FileServer(http.Dir("/")))

	readyCh := make(chan struct{})
	w := webview.New(false)
	v := &view{
		w:        w,
		animCh:   make(chan animation, 1),
		filePort: ln.Addr().(*net.TCPAddr).Port,
		readyCh:  readyCh,
	}

	win := w.Window()
	SetupWindow(win)

	w.Bind("moveWindow", func(dx, dy float64) {
		MoveWindow(win, dx, dy)
	})

	var once sync.Once
	w.Bind("_viewReady", func() {
		once.Do(func() { close(readyCh) })
	})

	w.SetHtml(viewHTML)
	return v
}

func (v *view) Show(anim animation) {
	select {
	case v.animCh <- anim:
	default:
	}
}

func (v *view) ShowWindow() {
	win := v.w.Window()
	v.w.Dispatch(func() { ShowWindow(win) })
}

func (v *view) HideWindow() {
	win := v.w.Window()
	v.w.Dispatch(func() { HideWindow(win) })
}

func (v *view) Run(ctx context.Context) {
	defer v.w.Destroy()
	go v.dispatch(ctx)
	go func() {
		<-ctx.Done()
		v.w.Dispatch(func() { v.w.Terminate() })
	}()
	v.w.Run()
}

func (v *view) dispatch(ctx context.Context) {
	select {
	case <-v.readyCh:
	case <-ctx.Done():
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case anim := <-v.animCh:
			v.w.Dispatch(func() {
				windowHeight := anim.windowHeight
				if windowHeight == 0 {
					windowHeight = anim.height
				}
				if anim.width > 0 && windowHeight > 0 {
					v.w.SetSize(anim.width, windowHeight, webview.HintNone)
					SetupWindow(v.w.Window())
				}
				url := fmt.Sprintf("http://127.0.0.1:%d%s", v.filePort, anim.filePath)
				v.w.Eval(fmt.Sprintf("showAnimation(%q, %d, %d, %d)",
					url, anim.fps, anim.width, anim.height))
			})
		}
	}
}

const viewHTML = `<!DOCTYPE html>
<html>
<head>
<style>
html, body {
  margin: 0;
  padding: 0;
  background: transparent;
  overflow: hidden;
  cursor: default;
  user-select: none;
}
canvas {
  display: block;
  image-rendering: pixelated;
}
</style>
</head>
<body>
<canvas id="sprite"></canvas>
<script>
var animTimer = null;

function showAnimation(src, fps, frameWidth, frameHeight) {
  var canvas = document.getElementById('sprite');
  var ctx = canvas.getContext('2d');
  var img = new Image();
  img.onload = function() {
    var cols = Math.floor(img.width / frameWidth);
    var rows = Math.floor(img.height / frameHeight);
    var frameCount = cols * rows;
    canvas.width = frameWidth;
    canvas.height = frameHeight;
    var frame = 0;
    if (animTimer) clearInterval(animTimer);
    animTimer = setInterval(function() {
      var col = frame % cols;
      var row = Math.floor(frame / cols);
      ctx.clearRect(0, 0, frameWidth, frameHeight);
      ctx.drawImage(img, col * frameWidth, row * frameHeight,
                    frameWidth, frameHeight,
                    0, 0, frameWidth, frameHeight);
      frame = (frame + 1) % frameCount;
    }, 1000 / fps);
  };
  img.src = src;
}

var dragging = false, lastX, lastY;
document.addEventListener('mousedown', function(e) {
  dragging = true; lastX = e.screenX; lastY = e.screenY;
});
document.addEventListener('mousemove', function(e) {
  if (!dragging) return;
  window.moveWindow(e.screenX - lastX, e.screenY - lastY);
  lastX = e.screenX; lastY = e.screenY;
});
document.addEventListener('mouseup', function() { dragging = false; });

window._viewReady();
</script>
</body>
</html>`
