package internal

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

type animation struct {
	filePath    string
	fps         int
	frame       dimension
	name        string
	description string
	window      dimension
}

type dimension struct {
	width  int
	height int
}

// live returns an error if the file cannot be opened or decoded as an image.
func (a animation) live() error {
	path, err := a.absFilePath()
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

// absFilePath returns the absolute form of a.filePath.
func (a animation) absFilePath() (string, error) {
	return filepath.Abs(a.filePath)
}

// windowDim returns the display window size: a.window if set, a.frame if set,
// otherwise the full image dimensions.
func (a animation) windowDim() dimension {
	if a.window.width > 0 || a.window.height > 0 {
		return a.window
	}
	if a.frame.width > 0 || a.frame.height > 0 {
		return a.frame
	}
	path, err := a.absFilePath()
	if err != nil {
		return dimension{}
	}
	f, err := os.Open(path)
	if err != nil {
		return dimension{}
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return dimension{}
	}
	return dimension{width: cfg.Width, height: cfg.Height}
}

// calcFps returns a.fps when set, otherwise infers fps from the number of
// frames in the spritesheet (cols × rows) so one full cycle takes one second.
// Returns an error when the animation's fields are insufficient to compute fps.
func (a animation) calcFps() (int, error) {
	if a.fps > 0 {
		return a.fps, nil
	}
	if a.frame.width == 0 || a.frame.height == 0 {
		return 0, fmt.Errorf("frame dimensions not set")
	}
	path, err := a.absFilePath()
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
	n := (cfg.Width / a.frame.width) * (cfg.Height / a.frame.height)
	if n == 0 {
		return 0, fmt.Errorf("computed zero frames from image %dx%d with frame %dx%d",
			cfg.Width, cfg.Height, a.frame.width, a.frame.height)
	}
	return n, nil
}
