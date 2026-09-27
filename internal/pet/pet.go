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
	frame       Dimension
	Name        string
	Description string
	window      Dimension
}

func (a Animation) live() error {
	_, err := a.imageSize()
	return err
}

func (a Animation) AbsFilePath() (string, error) {
	return filepath.Abs(a.filePath)
}

func (a Animation) FileSchemaPath() (string, error) {
	p, err := filepath.Abs(a.filePath)
	if err != nil {
		return "", err
	}
	return "file://" + p, nil
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

// Window returns the OS window dimensions for this animation.
// If one dimension is configured, the other is inferred from the frame's
// aspect ratio. If both are configured, the frame is scaled to fill the
// window (preserving aspect ratio), and the window is shrunk to that size.
func (a Animation) Window() Dimension {
	return a.window.Constrain(a.frame)
}

// DisplayFrame returns the dimensions at which each spritesheet frame is
// rendered. The frame always fills the window, so this equals Window().
func (a Animation) DisplayFrame() Dimension {
	return a.Window()
}

// FrameSize returns the spritesheet frame dimensions.
func (a Animation) FrameSize() Dimension {
	return a.frame
}

// CalcFps returns a.fps when set, otherwise infers fps from the number of
// frames in the spritesheet (cols × rows) so one full cycle takes one second.
// Returns an error when the animation's fields are insufficient to compute fps.
func (a Animation) CalcFps() (int, error) {
	if a.fps > 0 {
		return a.fps, nil
	}
	if !a.frame.Complete() {
		return 0, fmt.Errorf("frame dimensions not set")
	}
	img, err := a.imageSize()
	if err != nil {
		return 0, err
	}
	n := (img.Width / a.frame.Width) * (img.Height / a.frame.Height)
	if n == 0 {
		return 0, fmt.Errorf("computed zero frames from image %dx%d with frame %dx%d",
			img.Width, img.Height, a.frame.Width, a.frame.Height)
	}
	return n, nil
}

func resolveAnimation(a config.AnimationConfig, p config.PetConfig, cfg config.Config) Animation {
	return Animation{
		filePath: a.File,
		fps:      coalesce(a.FPS, p.FPS, cfg.FPS),
		frame: Dimension{
			Height: coalesce(a.Frame.Height, p.Frame.Height, cfg.Frame.Height),
			Width:  coalesce(a.Frame.Width, p.Frame.Width, cfg.Frame.Width),
		},
		Name:        a.Name,
		Description: a.Description,
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
