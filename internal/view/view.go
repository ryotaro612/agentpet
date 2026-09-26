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

	"github.com/ryotaro612/agentpet/internal/pet"
	"github.com/webview/webview"
)

//go:embed view.html
var viewHTMLSrc string

// View wraps the native webview window and drives spritesheet animations.
type View struct {
	w        webview.WebView
	animCh   chan pet.Animation
	readyCh  chan struct{}
	logger   *slog.Logger
	htmlPath string
	// shown is set after the first animation is displayed and is accessed only
	// from Dispatch callbacks, which run on the webview main thread.
	shown       bool
	stateMu     sync.RWMutex
	currentPet  string
	currentAnim string
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
	PreventTerminateOnHide()
	v := &View{
		w:        w,
		animCh:   make(chan pet.Animation, 1),
		readyCh:  readyCh,
		logger:   logger,
		htmlPath: f.Name(),
	}

	win := w.Window()
	SetupWindow(win)
	PreventHide(win)
	SetupQuit()

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

// Show records the current pet and animation, then queues the animation for
// display, dropping the previous one if unread.
func (v *View) Show(petName string, anim pet.Animation) {
	v.stateMu.Lock()
	v.currentPet = petName
	v.currentAnim = anim.Name
	v.stateMu.Unlock()
	select {
	case v.animCh <- anim:
	default:
	}
}

// Current returns the pet name and animation name that were most recently shown.
func (v *View) Current() (petName, animName string) {
	v.stateMu.RLock()
	defer v.stateMu.RUnlock()
	return v.currentPet, v.currentAnim
}

// ShowWindow brings the pet window to the front.
func (v *View) ShowWindow() {
	if v.w == nil {
		return
	}
	win := v.w.Window()
	v.w.Dispatch(func() { ShowWindow(win) })
}

// HideWindow hides the pet window.
func (v *View) HideWindow() {
	if v.w == nil {
		return
	}
	win := v.w.Window()
	v.w.Dispatch(func() { HideWindow(win) })
}

// Noop returns a View that accepts method calls without doing anything.
// Intended for use in tests.
func Noop() *View {
	return &View{animCh: make(chan pet.Animation, 1)}
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
				window, displayFrame := anim.Layout()
				if window.NonZero() {
					if !v.shown {
						SetWindowSize(v.w.Window(), window.Width, window.Height)
						v.shown = true
					} else {
						ResizeWindowKeepingPosition(v.w.Window(), window.Width, window.Height)
					}
				}
				fps, fpsErr := anim.CalcFps()
				if fpsErr != nil {
					v.logger.Warn("skipping animation: cannot compute fps",
						"animation", anim.Name, "err", fpsErr)
					return
				}
				filePath, _ := anim.AbsFilePath()
				if filePath == "" {
					return
				}
				v.w.Eval(fmt.Sprintf("showAnimation(%q, %d, %d, %d, %d, %d)",
					"file://"+filePath, fps,
					anim.Frame.Width, anim.Frame.Height,
					displayFrame.Width, displayFrame.Height))
			})
		}
	}
}
