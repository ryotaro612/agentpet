package view

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	_ "embed"

	"github.com/webview/webview"
)

//go:embed view.html
var viewHTMLSrc string

// Anim carries pre-computed display parameters for a single spritesheet animation.
type Anim struct {
	Name    string
	FPS     int
	FPSErr  error
	FrameW  int
	FrameH  int
	WinW    int
	WinH    int
}

// View wraps the native webview window and drives spritesheet animations.
type View struct {
	w          webview.WebView
	animCh     chan Anim
	readyCh    chan struct{}
	serverAddr atomic.Pointer[string]
	logger     *slog.Logger
}

func New(logger *slog.Logger) *View {
	tmpl := template.Must(template.New("view").Parse(viewHTMLSrc))
	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		panic("view: failed to render HTML template: " + err.Error())
	}
	html := buf.String()

	readyCh := make(chan struct{})
	w := webview.New(false)
	v := &View{
		w:       w,
		animCh:  make(chan Anim, 1),
		readyCh: readyCh,
		logger:  logger,
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

	w.SetHtml(html)
	return v
}

// SetServerAddr tells the view which base URL to use for /image requests.
func (v *View) SetServerAddr(addr string) {
	v.serverAddr.Store(&addr)
}

// Show queues an animation for display, dropping the previous one if unread.
func (v *View) Show(anim Anim) {
	select {
	case v.animCh <- anim:
	default:
	}
}

// ShowWindow brings the pet window to the front.
func (v *View) ShowWindow() {
	win := v.w.Window()
	v.w.Dispatch(func() { ShowWindow(win) })
}

// HideWindow hides the pet window.
func (v *View) HideWindow() {
	win := v.w.Window()
	v.w.Dispatch(func() { HideWindow(win) })
}

// Run starts the webview event loop. It blocks until ctx is cancelled or the
// window is closed.
func (v *View) Run(ctx context.Context) {
	defer v.w.Destroy()
	go v.dispatch(ctx)
	go func() {
		<-ctx.Done()
		v.w.Dispatch(func() { v.w.Terminate() })
	}()
	v.w.Run()
}

func (v *View) dispatch(ctx context.Context) {
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
				if anim.WinW > 0 && anim.WinH > 0 {
					v.w.SetSize(anim.WinW, anim.WinH, webview.HintNone)
					SetupWindow(v.w.Window())
				}
				if anim.FPSErr != nil {
					v.logger.Warn("skipping animation: cannot compute fps",
						"animation", anim.Name, "err", anim.FPSErr)
					return
				}
				p := v.serverAddr.Load()
				if p == nil {
					return
				}
				url := "http://" + *p + "/image"
				v.w.Eval(fmt.Sprintf("showAnimation(%q, %d, %d, %d)",
					url, anim.FPS, anim.FrameW, anim.FrameH))
			})
		}
	}
}
