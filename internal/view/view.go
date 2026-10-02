package view

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/ryotaro612/agentpet/internal/pet"
)

//go:embed view.html
var viewHTMLSrc string

// View wraps the native webview window and drives spritesheet animations.
type View struct {
	w           internalView
	animCh      chan pet.Animation
	logger      *slog.Logger
	htmlPath    string
	stateMu     sync.RWMutex
	currentPet  string
	currentAnim string
}

func New(w internalView, logger *slog.Logger) (*View, error) {
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

	v := &View{
		w:        w,
		animCh:   make(chan pet.Animation, 1),
		logger:   logger,
		htmlPath: f.Name(),
	}

	v.w.PreventTerminateOnHide()
	v.w.SetUpWIndow()
	v.w.PreventHide()
	v.w.SetupQuit()

	v.w.Bind("moveWindow", func(dx, dy float64) {
		v.w.MoveWindow(dx, dy)
	})

	v.w.Navigate("file://" + f.Name())
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

// CurrentPet returns the name of the pet that was most recently shown.
func (v *View) CurrentPet() string {
	v.stateMu.RLock()
	defer v.stateMu.RUnlock()
	return v.currentPet
}

// CurrentAnimation returns the name of the animation that was most recently shown.
func (v *View) CurrentAnimation() string {
	v.stateMu.RLock()
	defer v.stateMu.RUnlock()
	return v.currentAnim
}

// ShowWindow brings the pet window to the front.
func (v *View) ShowWindow() {
	v.w.SHowWindow()
}

// HideWindow hides the pet window.
func (v *View) HideWindow() {
	v.w.HideWindow()
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
					v.w.SetWindowSize(win.Width, win.Height)
					shown = true
				} else {
					v.w.ResizeWindowKeepingPosition(win.Width, win.Height)
				}
				fps, fpsErr := anim.CalcFps()
				if fpsErr != nil {
					v.logger.Warn("skipping animation: cannot compute fps",
						"animation", anim.Name, "err", fpsErr)
					return
				}
				count, countErr := anim.Count()
				if countErr != nil {
					v.logger.Warn("skipping animation: cannot compute frame count",
						"animation", anim.Name, "err", countErr)
					return
				}
				filePath, err := anim.FileSchemaPath()
				if err != nil {
					v.logger.Warn("skipping animation: cannot resolve path",
						"animation", anim.Name, "err", err)
					return
				}
				frame := anim.FrameSize()
				v.w.Eval(fmt.Sprintf("showAnimation(%q, %d, %d, %d, %d, %d, %d)",
					filePath, fps, frame.Width, frame.Height, win.Width, win.Height, count))
			})
		}
	}
}
