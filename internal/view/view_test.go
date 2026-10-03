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

	"github.com/ryotaro612/agentpet/internal/config"
	"github.com/ryotaro612/agentpet/internal/pet"
)

func TestViewCurrent(t *testing.T) {
	t.Parallel()

	t.Run("returns the most recently shown pet name and animation name", func(t *testing.T) {
		t.Parallel()
		v := &View{animCh: make(chan pet.Animation, 1)}
		v.Show("cat", pet.Animation{Name: "idle"})
		if got := v.CurrentPet(); got != "cat" {
			t.Errorf("CurrentPet() = %q, want %q", got, "cat")
		}
		if got := v.CurrentAnimation(); got != "idle" {
			t.Errorf("CurrentAnimation() = %q, want %q", got, "idle")
		}
	})
}

func TestViewDispatch(t *testing.T) {
	t.Parallel()

	t.Run("calls SetWindowSize on the first animation", func(t *testing.T) {
		t.Parallel()
		mock := NewMockInternalView()
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
		<-mock.DispatchDone
		cancel()
		<-done

		if len(mock.SetWindowSizeCalls) != 1 {
			t.Errorf("SetWindowSize called %d time(s), want 1", len(mock.SetWindowSizeCalls))
		}
		if len(mock.ResizeWindowCalls) != 0 {
			t.Errorf("ResizeWindowKeepingPosition called %d time(s), want 0", len(mock.ResizeWindowCalls))
		}
		if len(mock.EvalCalls) != 1 {
			t.Errorf("Eval called %d time(s), want 1", len(mock.EvalCalls))
		}
	})

	t.Run("calls ResizeWindowKeepingPosition on subsequent animations", func(t *testing.T) {
		t.Parallel()
		mock := NewMockInternalView()
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
		<-mock.DispatchDone
		v.animCh <- anim
		<-mock.DispatchDone
		cancel()
		<-done

		if len(mock.SetWindowSizeCalls) != 1 {
			t.Errorf("SetWindowSize called %d time(s), want 1", len(mock.SetWindowSizeCalls))
		}
		if len(mock.ResizeWindowCalls) != 1 {
			t.Errorf("ResizeWindowKeepingPosition called %d time(s), want 1", len(mock.ResizeWindowCalls))
		}
	})

	t.Run("given a valid animation, adjusts the window size and calls showAnimation with the correct arguments", func(t *testing.T) {
		t.Parallel()
		mock := NewMockInternalView()
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
		<-mock.DispatchDone
		v.animCh <- anim
		<-mock.DispatchDone
		cancel()
		<-done

		absPath, err := filepath.Abs(spritePath)
		if err != nil {
			t.Fatal(err)
		}
		// No window configured: the window matches the sprite frame (32×32).
		wantDim := [2]int{32, 32}
		if len(mock.SetWindowSizeCalls) != 1 || mock.SetWindowSizeCalls[0] != wantDim {
			t.Errorf("window size on first play = %v, want [%v]", mock.SetWindowSizeCalls, wantDim)
		}
		if len(mock.ResizeWindowCalls) != 1 || mock.ResizeWindowCalls[0] != wantDim {
			t.Errorf("window size on second play = %v, want [%v]", mock.ResizeWindowCalls, wantDim)
		}
		wantEval := fmt.Sprintf("showAnimation(%q, 2, 32, 32, 32, 32, 2)", "file://"+absPath)
		for i, got := range mock.EvalCalls {
			if got != wantEval {
				t.Errorf("animation rendered on play %d = %q, want %q", i+1, got, wantEval)
			}
		}
		if len(mock.EvalCalls) != 2 {
			t.Errorf("animation rendered %d time(s) across two plays, want 2", len(mock.EvalCalls))
		}
	})

	t.Run("passes the configured frame count to showAnimation", func(t *testing.T) {
		t.Parallel()
		mock := NewMockInternalView()
		v := newTestView(mock)
		readyCh := make(chan struct{})
		close(readyCh)
		spritePath := writeSpritePNG(t, 128, 32)
		anim := pet.NewKeeper(config.Config{
			Pets: []config.PetConfig{{
				Name:  "test",
				Frame: config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{{
					Name: "idle", File: spritePath, FPS: 10, Count: 2,
				}},
			}},
		}).Animations("test")[0]

		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan struct{})
		go func() {
			v.dispatch(ctx, readyCh)
			close(done)
		}()
		v.animCh <- anim
		<-mock.DispatchDone
		cancel()
		<-done

		absPath, err := filepath.Abs(spritePath)
		if err != nil {
			t.Fatal(err)
		}
		want := fmt.Sprintf("showAnimation(%q, 10, 32, 32, 32, 32, 2)", "file://"+absPath)
		if len(mock.EvalCalls) != 1 || mock.EvalCalls[0] != want {
			t.Errorf("Eval calls = %q, want [%q]", mock.EvalCalls, want)
		}
	})

}

// animForTest returns an Animation with explicit fps so CalcFps succeeds
// without needing a real image on disk. fps and frame are set through the
// config+keeper pipeline because Animation.fps is unexported.
func animForTest(fps int, frame pet.Dimension) pet.Animation {
	anims := pet.NewKeeper(config.Config{
		Pets: []config.PetConfig{{
			Name:       "test",
			FPS:        fps,
			Frame:      config.DimensionConfig{Width: frame.Width, Height: frame.Height},
			Window:     config.DimensionConfig{Width: frame.Width, Height: frame.Height},
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

func newTestView(m *MockInternalView) *View {
	return &View{
		w:      m,
		animCh: make(chan pet.Animation, 1),
		logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
}
