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
	path, err := a.AbsFilePath()
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, _, err = image.DecodeConfig(f)
	return err
}

func (a Animation) AbsFilePath() (string, error) {
	return filepath.Abs(a.filePath)
}

// WindowDim returns the display window size: a.window if set, a.Frame if set,
// otherwise the full image dimensions.
func (a Animation) WindowDim() Dimension {
	if a.window.NonZero() {
		return a.window
	}
	if a.Frame.NonZero() {
		return a.Frame
	}
	path, err := a.AbsFilePath()
	if err != nil {
		return Dimension{}
	}
	f, err := os.Open(path)
	if err != nil {
		return Dimension{}
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return Dimension{}
	}
	return Dimension{Width: cfg.Width, Height: cfg.Height}
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
	path, err := a.AbsFilePath()
	if err != nil {
		return 0, err
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, err
	}
	n := (cfg.Width / a.Frame.Width) * (cfg.Height / a.Frame.Height)
	if n == 0 {
		return 0, fmt.Errorf("computed zero frames from image %dx%d with frame %dx%d",
			cfg.Width, cfg.Height, a.Frame.Width, a.Frame.Height)
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
