package internal

import (
	"context"
	"fmt"

	"github.com/webview/webview"
)

type view struct {
	w      webview.WebView
	animCh chan animation
}

func NewView() *view {
	w := webview.New(true)
	v := &view{
		w:      w,
		animCh: make(chan animation, 1),
	}
	win := w.Window()
	w.Bind("moveWindow", func(dx, dy float64) {
		MoveWindow(win, dx, dy)
	})
	w.SetHtml(viewHTML)
	HideToolbar(win)
	MakeTransparent(win)
	return v
}

func (v *view) Show(anim animation) {
	select {
	case v.animCh <- anim:
	default:
	}
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
				}
				v.w.Eval(fmt.Sprintf("showAnimation(%q, %d, %d, %d)",
					anim.filePath, anim.fps, anim.width, anim.height))
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
    var frameCount = Math.floor(img.width / frameWidth);
    canvas.width = frameWidth;
    canvas.height = frameHeight;
    var frame = 0;
    if (animTimer) clearInterval(animTimer);
    animTimer = setInterval(function() {
      ctx.clearRect(0, 0, frameWidth, frameHeight);
      ctx.drawImage(img, frame * frameWidth, 0, frameWidth, frameHeight,
                    0, 0, frameWidth, frameHeight);
      frame = (frame + 1) % frameCount;
    }, 1000 / fps);
  };
  img.src = 'file://' + src;
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
</script>
</body>
</html>`
