package view

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"text/template"
	_ "embed"

	"github.com/webview/webview"
)

//go:embed view.html
var viewHTMLSrc string

// Anim carries pre-computed display parameters for a single spritesheet animation.
type Anim struct {
	Name     string
	FPS      int
	FPSErr   error
	FrameW   int
	FrameH   int
	WinW     int
	WinH     int
	FilePath string
}

// View wraps the native webview window and drives spritesheet animations.
type View struct {
	w        webview.WebView
	animCh   chan Anim
	readyCh  chan struct{}
	logger   *slog.Logger
	htmlPath string
}

func New(logger *slog.Logger) *View {
	tmpl := template.Must(template.New("view").Parse(viewHTMLSrc))
	var buf strings.Builder
	if err := tmpl.Execute(&buf, nil); err != nil {
		panic("view: failed to render HTML template: " + err.Error())
	}

	f, err := os.CreateTemp("", "agentpet-*.html")
	if err != nil {
		panic("view: failed to create temp HTML file: " + err.Error())
	}
	if _, err := f.WriteString(buf.String()); err != nil {
		panic("view: failed to write HTML: " + err.Error())
	}
	f.Close()

	readyCh := make(chan struct{})
	w := webview.New(false)
	v := &View{
		w:        w,
		animCh:   make(chan Anim, 1),
		readyCh:  readyCh,
		logger:   logger,
		htmlPath: f.Name(),
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

	w.Navigate("file://" + f.Name())
	return v
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
	defer os.Remove(v.htmlPath)
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
				if anim.FilePath == "" {
					return
				}
				url := "file://" + anim.FilePath
				v.w.Eval(fmt.Sprintf("showAnimation(%q, %d, %d, %d)",
					url, anim.FPS, anim.FrameW, anim.FrameH))
			})
		}
	}
}
