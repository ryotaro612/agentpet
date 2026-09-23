package pet

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/ryotaro612/agentpet/internal/config"
)

type Animation struct {
	filePath    string
	fps         int
	Frame       Dimension
	Name        string
	description string
	window      Dimension
}

type Dimension struct {
	Width  int
	Height int
}

func (d Dimension) NonZero() bool {
	return d.Width > 0 || d.Height > 0
}

func (a Animation) live() error {
	_, err := a.imageSize()
	return err
}

func (a Animation) AbsFilePath() (string, error) {
	return filepath.Abs(a.filePath)
}

func (a Animation) imageSize() (Dimension, error) {
	path, err := a.AbsFilePath()
	if err != nil {
		return Dimension{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return Dimension{}, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return Dimension{}, err
	}
	return Dimension{Width: cfg.Width, Height: cfg.Height}, nil
}

// Layout returns the OS window size and the display frame size for rendering.
// If a.window has exactly one dimension set, the other is inferred from the
// image's aspect ratio. If the window is smaller than a.Frame in either
// dimension, the display frame is scaled down to fit within the window while
// preserving the frame's aspect ratio.
func (a Animation) Layout() (window, displayFrame Dimension) {
	w, h := a.window.Width, a.window.Height
	switch {
	case w > 0 && h > 0:
		window = a.window
	case w > 0 || h > 0:
		if img, err := a.imageSize(); err == nil && img.Width > 0 && img.Height > 0 {
			if w > 0 {
				h = w * img.Height / img.Width
			} else {
				w = h * img.Width / img.Height
			}
		}
		window = Dimension{Width: w, Height: h}
	default:
		if a.Frame.NonZero() {
			window = a.Frame
		} else {
			window, _ = a.imageSize()
		}
	}

	if !a.Frame.NonZero() {
		displayFrame = window
		return
	}
	if !window.NonZero() || (window.Width >= a.Frame.Width && window.Height >= a.Frame.Height) {
		displayFrame = a.Frame
		return
	}
	scaleW := float64(window.Width) / float64(a.Frame.Width)
	scaleH := float64(window.Height) / float64(a.Frame.Height)
	scale := min(scaleW, scaleH)
	displayFrame = Dimension{
		Width:  int(float64(a.Frame.Width) * scale),
		Height: int(float64(a.Frame.Height) * scale),
	}
	return
}

// CalcFps returns a.fps when set, otherwise infers fps from the number of
// frames in the spritesheet (cols × rows) so one full cycle takes one second.
// Returns an error when the animation's fields are insufficient to compute fps.
func (a Animation) CalcFps() (int, error) {
	if a.fps > 0 {
		return a.fps, nil
	}
	if a.Frame.Width == 0 || a.Frame.Height == 0 {
		return 0, fmt.Errorf("frame dimensions not set")
	}
	img, err := a.imageSize()
	if err != nil {
		return 0, err
	}
	n := (img.Width / a.Frame.Width) * (img.Height / a.Frame.Height)
	if n == 0 {
		return 0, fmt.Errorf("computed zero frames from image %dx%d with frame %dx%d",
			img.Width, img.Height, a.Frame.Width, a.Frame.Height)
	}
	return n, nil
}

func resolveAnimation(a config.AnimationConfig, p config.PetConfig, cfg config.Config) Animation {
	return Animation{
		filePath: a.File,
		fps:      coalesce(a.FPS, p.FPS, cfg.FPS),
		Frame: Dimension{
			Height: coalesce(a.Frame.Height, p.Frame.Height, cfg.Frame.Height),
			Width:  coalesce(a.Frame.Width, p.Frame.Width, cfg.Frame.Width),
		},
		Name:        a.Name,
		description: a.Description,
		window: Dimension{
			Height: coalesce(a.Window.Height, p.Window.Height, cfg.Window.Height),
			Width:  coalesce(a.Window.Width, p.Window.Width, cfg.Window.Width),
		},
	}
}

func coalesce(vals ...int) int {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}
