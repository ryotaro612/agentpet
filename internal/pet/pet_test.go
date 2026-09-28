package pet

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnimationWindow(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		anim Animation
		want Dimension
	}{
		{
			"returns Frame when no window is configured",
			Animation{frame: Dimension{Width: 32, Height: 32}},
			Dimension{Width: 32, Height: 32},
		},
		{
			"infers height from Frame aspect ratio when only window width is set",
			Animation{frame: Dimension{Width: 200, Height: 100}, window: Dimension{Width: 100}},
			Dimension{Width: 100, Height: 50},
		},
		{
			"infers width from Frame aspect ratio when only window height is set",
			Animation{frame: Dimension{Width: 200, Height: 100}, window: Dimension{Height: 50}},
			Dimension{Width: 100, Height: 50},
		},
		{
			"scales Frame to fit the configured window preserving aspect ratio",
			Animation{frame: Dimension{Width: 200, Height: 100}, window: Dimension{Width: 100, Height: 100}},
			Dimension{Width: 100, Height: 50},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := c.anim.Window(); got != c.want {
				t.Errorf("Window() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestAnimationCalcFps(t *testing.T) {
	t.Parallel()

	t.Run("returns explicit fps when set", func(t *testing.T) {
		t.Parallel()
		anim := Animation{fps: 12, frame: Dimension{Width: 32, Height: 32}}
		got, err := anim.CalcFps()
		if err != nil {
			t.Fatalf("CalcFps() error = %v", err)
		}
		if got != 12 {
			t.Errorf("CalcFps() = %d, want 12", got)
		}
	})

	t.Run("returns an error when fps is zero and frame dimensions are not set", func(t *testing.T) {
		t.Parallel()
		anim := Animation{}
		if _, err := anim.CalcFps(); err == nil {
			t.Error("CalcFps() = nil, want error")
		}
	})

	t.Run("infers fps from spritesheet frame count when fps is not set", func(t *testing.T) {
		t.Parallel()
		// 64×32 image with 32×32 frames → 2 columns × 1 row = 2 frames → fps 2
		path := writePNG(t, 64, 32)
		anim := Animation{filePath: path, frame: Dimension{Width: 32, Height: 32}}
		got, err := anim.CalcFps()
		if err != nil {
			t.Fatalf("CalcFps() error = %v", err)
		}
		if got != 2 {
			t.Errorf("CalcFps() = %d, want 2", got)
		}
	})
}

func TestAnimationFileSchemaPath(t *testing.T) {
	t.Parallel()
	t.Run("returns a file:// URL containing the absolute path", func(t *testing.T) {
		t.Parallel()
		anim := Animation{filePath: "sprite.png"}
		got, err := anim.FileSchemaPath()
		if err != nil {
			t.Fatalf("FileSchemaPath() error = %v", err)
		}
		if !strings.HasPrefix(got, "file://") {
			t.Errorf("FileSchemaPath() = %q, want prefix \"file://\"", got)
		}
		abs, _ := filepath.Abs("sprite.png")
		if want := "file://" + abs; got != want {
			t.Errorf("FileSchemaPath() = %q, want %q", got, want)
		}
	})

	t.Run("expands ~/ to the home directory", func(t *testing.T) {
		t.Parallel()
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}
		anim := Animation{filePath: "~/sprites/cat.png"}
		got, err := anim.FileSchemaPath()
		if err != nil {
			t.Fatalf("FileSchemaPath() error = %v", err)
		}
		want := "file://" + filepath.Join(home, "sprites", "cat.png")
		if got != want {
			t.Errorf("FileSchemaPath() = %q, want %q", got, want)
		}
	})
}

func writePNG(t *testing.T, w, h int) string {
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
