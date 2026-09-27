package view

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ryotaro612/agentpet/internal/config"
	"github.com/ryotaro612/agentpet/internal/pet"
)

func TestViewCurrent(t *testing.T) {
	t.Parallel()

	t.Run("returns the most recently shown pet name and animation name", func(t *testing.T) {
		t.Parallel()
		v := &View{animCh: make(chan pet.Animation, 1)}
		v.Show("cat", pet.Animation{Name: "idle"})
		got := v.Current()
		if got.Pet != "cat" {
			t.Errorf("Current().Pet = %q, want %q", got.Pet, "cat")
		}
		if got.Animation != "idle" {
			t.Errorf("Current().Animation = %q, want %q", got.Animation, "idle")
		}
	})
}

func TestViewDispatch(t *testing.T) {
	t.Parallel()

	t.Run("exits immediately when context is cancelled before ready", func(t *testing.T) {
		t.Parallel()
		mock := newMockView()
		v := newTestView(mock)
		readyCh := make(chan struct{}) // never closed

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		done := make(chan struct{})
		go func() {
			v.dispatch(ctx, readyCh)
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("dispatch did not exit after context cancellation")
		}
		if len(mock.evalCalls) != 0 {
			t.Errorf("dispatch processed animations after cancellation: evalCalls = %v", mock.evalCalls)
		}
	})

	t.Run("calls SetWindowSize on the first animation", func(t *testing.T) {
		t.Parallel()
		mock := newMockView()
		v := newTestView(mock)
		readyCh := make(chan struct{})
		close(readyCh)

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan struct{})
		go func() {
			v.dispatch(ctx, readyCh)
			close(done)
		}()

		v.animCh <- animForTest(8, pet.Dimension{Width: 32, Height: 32})
		<-mock.dispatchDone
		cancel()
		<-done

		if len(mock.setWindowSizeCalls) != 1 {
			t.Errorf("SetWindowSize called %d time(s), want 1", len(mock.setWindowSizeCalls))
		}
		if len(mock.resizeWindowCalls) != 0 {
			t.Errorf("ResizeWindowKeepingPosition called %d time(s), want 0", len(mock.resizeWindowCalls))
		}
		if len(mock.evalCalls) != 1 {
			t.Errorf("Eval called %d time(s), want 1", len(mock.evalCalls))
		}
	})

	t.Run("calls ResizeWindowKeepingPosition on subsequent animations", func(t *testing.T) {
		t.Parallel()
		mock := newMockView()
		v := newTestView(mock)
		readyCh := make(chan struct{})
		close(readyCh)

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan struct{})
		go func() {
			v.dispatch(ctx, readyCh)
			close(done)
		}()

		anim := animForTest(8, pet.Dimension{Width: 32, Height: 32})
		v.animCh <- anim
		<-mock.dispatchDone
		v.animCh <- anim
		<-mock.dispatchDone
		cancel()
		<-done

		if len(mock.setWindowSizeCalls) != 1 {
			t.Errorf("SetWindowSize called %d time(s), want 1", len(mock.setWindowSizeCalls))
		}
		if len(mock.resizeWindowCalls) != 1 {
			t.Errorf("ResizeWindowKeepingPosition called %d time(s), want 1", len(mock.resizeWindowCalls))
		}
	})

	t.Run("given a valid animation, adjusts the window size and calls showAnimation with the correct arguments", func(t *testing.T) {
		t.Parallel()
		mock := newMockView()
		v := newTestView(mock)
		readyCh := make(chan struct{})
		close(readyCh)

		// 64×32 sprite sheet with 32×32 frames → 2 columns × 1 row = 2 fps (inferred).
		spritePath := writeSpritePNG(t, 64, 32)
		anim := pet.NewKeeper(config.Config{
			Pets: []config.PetConfig{{
				Name:       "test",
				Frame:      config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{{Name: "idle", File: spritePath}},
			}},
		}).Animations("test")[0]

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan struct{})
		go func() {
			v.dispatch(ctx, readyCh)
			close(done)
		}()

		v.animCh <- anim
		<-mock.dispatchDone
		v.animCh <- anim
		<-mock.dispatchDone
		cancel()
		<-done

		absPath, err := filepath.Abs(spritePath)
		if err != nil {
			t.Fatal(err)
		}
		// No window configured: the window matches the sprite frame (32×32).
		wantDim := [2]int{32, 32}
		if len(mock.setWindowSizeCalls) != 1 || mock.setWindowSizeCalls[0] != wantDim {
			t.Errorf("window size on first play = %v, want [%v]", mock.setWindowSizeCalls, wantDim)
		}
		if len(mock.resizeWindowCalls) != 1 || mock.resizeWindowCalls[0] != wantDim {
			t.Errorf("window size on second play = %v, want [%v]", mock.resizeWindowCalls, wantDim)
		}
		wantEval := fmt.Sprintf("showAnimation(%q, 2, 32, 32, 32, 32)", "file://"+absPath)
		for i, got := range mock.evalCalls {
			if got != wantEval {
				t.Errorf("animation rendered on play %d = %q, want %q", i+1, got, wantEval)
			}
		}
		if len(mock.evalCalls) != 2 {
			t.Errorf("animation rendered %d time(s) across two plays, want 2", len(mock.evalCalls))
		}
	})

	t.Run("does not call Eval when CalcFps fails", func(t *testing.T) {
		t.Parallel()
		mock := newMockView()
		v := newTestView(mock)
		readyCh := make(chan struct{})
		close(readyCh)

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan struct{})
		go func() {
			v.dispatch(ctx, readyCh)
			close(done)
		}()

		// Zero-value Animation has no fps and zero Frame, so CalcFps returns an error.
		v.animCh <- pet.Animation{}
		<-mock.dispatchDone
		cancel()
		<-done

		if len(mock.evalCalls) != 0 {
			t.Errorf("Eval called %d time(s), want 0 when CalcFps fails", len(mock.evalCalls))
		}
	})
}

// mockView records calls made by dispatch and executes Dispatch closures
// synchronously so tests can assert state after a known number of animations.
type mockView struct {
	setWindowSizeCalls [][2]int
	resizeWindowCalls  [][2]int
	evalCalls          []string
	// dispatchDone receives a value after each Dispatch closure completes.
	dispatchDone chan struct{}
}

func newMockView() *mockView {
	return &mockView{dispatchDone: make(chan struct{}, 16)}
}

func (m *mockView) PreventTerminateOnHide()    {}
func (m *mockView) SetUpWIndow()               {}
func (m *mockView) PreventHide()               {}
func (m *mockView) SetupQuit()                 {}
func (m *mockView) MoveWindow(_, _ float64)    {}
func (m *mockView) SHowWindow()                {}
func (m *mockView) HideWindow()                {}
func (m *mockView) Bind(_ string, _ any) error { return nil }
func (m *mockView) Navigate(_ string)          {}
func (m *mockView) Terminate()                 {}
func (m *mockView) Run()                       {}
func (m *mockView) Destroy()                   {}
func (m *mockView) SetWindowSize(w, h int) {
	m.setWindowSizeCalls = append(m.setWindowSizeCalls, [2]int{w, h})
}
func (m *mockView) ResizeWindowKeepingPosition(w, h int) {
	m.resizeWindowCalls = append(m.resizeWindowCalls, [2]int{w, h})
}
func (m *mockView) Eval(js string) { m.evalCalls = append(m.evalCalls, js) }
func (m *mockView) Dispatch(f func()) {
	f()
	m.dispatchDone <- struct{}{}
}

// animForTest returns an Animation with explicit fps so CalcFps succeeds
// without needing a real image on disk. fps and frame are set through the
// config+keeper pipeline because Animation.fps is unexported.
func animForTest(fps int, frame pet.Dimension) pet.Animation {
	anims := pet.NewKeeper(config.Config{
		Pets: []config.PetConfig{{
			Name: "test",
			FPS:  fps,
			Frame: config.DimensionConfig{
				Width:  frame.Width,
				Height: frame.Height,
			},
			Animations: []config.AnimationConfig{{Name: "idle", File: "idle.png"}},
		}},
	}).Animations("test")
	return anims[0]
}

func writeSpritePNG(t *testing.T, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	path := filepath.Join(t.TempDir(), "sprite.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return path
}

func newTestView(m *mockView) *View {
	return &View{
		w:      m,
		animCh: make(chan pet.Animation, 1),
		logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
}
