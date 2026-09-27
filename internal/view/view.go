package view

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
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
	logger   *slog.Logger
	htmlPath string
	stateMu     sync.RWMutex
	currentPet  string
	currentAnim string
}

func New(logger *slog.Logger) (*View, error) {
	f, err := os.CreateTemp("", "agentpet-*.html")
	if err != nil {
		return nil, fmt.Errorf("view: failed to create temp HTML file: %w", err)
	}
	if _, err := f.WriteString(viewHTMLSrc); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, fmt.Errorf("view: failed to write HTML: %w", err)
	}
	f.Close()

	w := webview.New(false)
	PreventTerminateOnHide()
	v := &View{
		w:        w,
		animCh:   make(chan pet.Animation, 1),
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

	w.Navigate("file://" + f.Name())
	return v, nil
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
	readyCh := make(chan struct{})
	var once sync.Once
	v.w.Bind("_viewReady", func() {
		once.Do(func() { close(readyCh) })
	})
	go v.dispatch(ctx, readyCh)
	go func() {
		<-ctx.Done()
		v.w.Dispatch(func() { v.w.Terminate() })
	}()
	v.w.Run()
}

func (v *View) dispatch(ctx context.Context, readyCh <-chan struct{}) {
	select {
	case <-readyCh:
	case <-ctx.Done():
		return
	}
	shown := false
	for {
		select {
		case <-ctx.Done():
			return
		case anim := <-v.animCh:
			v.w.Dispatch(func() {
				win := anim.Window()
				if !shown {
					SetWindowSize(v.w.Window(), win.Width, win.Height)
					shown = true
				} else {
					ResizeWindowKeepingPosition(v.w.Window(), win.Width, win.Height)
				}
				fps, fpsErr := anim.CalcFps()
				if fpsErr != nil {
					v.logger.Warn("skipping animation: cannot compute fps",
						"animation", anim.Name, "err", fpsErr)
					return
				}
				filePath, err := anim.FileSchemaPath()
				if err != nil {
					v.logger.Warn("skipping animation: cannot resolve path",
						"animation", anim.Name, "err", err)
					return
				}
				frame := anim.FrameSize()
				v.w.Eval(fmt.Sprintf("showAnimation(%q, %d, %d, %d, %d, %d)",
					filePath, fps, frame.Width, frame.Height, win.Width, win.Height))
			})
		}
	}
}
