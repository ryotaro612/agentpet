package pet

type Dimension struct {
	Width  int
	Height int
}

// Complete reports whether both dimensions are set.
func (d Dimension) Complete() bool { return d.Width > 0 && d.Height > 0 }

// HasWidth reports whether the width dimension is set.
func (d Dimension) HasWidth() bool { return d.Width > 0 }

// HasHeight reports whether the height dimension is set.
func (d Dimension) HasHeight() bool { return d.Height > 0 }

// ScaleTo scales d to fit within target while preserving d's aspect ratio.
func (d Dimension) ScaleTo(target Dimension) Dimension {
	scaleW := float64(target.Width) / float64(d.Width)
	scaleH := float64(target.Height) / float64(d.Height)
	scale := min(scaleW, scaleH)
	return Dimension{Width: int(float64(d.Width) * scale), Height: int(float64(d.Height) * scale)}
}

// WithWidth returns a Dimension with the given width and height inferred from d's aspect ratio.
func (d Dimension) WithWidth(w int) Dimension {
	return Dimension{Width: w, Height: w * d.Height / d.Width}
}

// WithHeight returns a Dimension with the given height and width inferred from d's aspect ratio.
func (d Dimension) WithHeight(h int) Dimension {
	return Dimension{Width: h * d.Width / d.Height, Height: h}
}

// Tiles returns the number of non-overlapping frame-sized cells that fit in d.
func (d Dimension) Tiles(frame Dimension) int {
	return (d.Width / frame.Width) * (d.Height / frame.Height)
}

// Constrain scales frame to fit within d while preserving frame's aspect ratio.
// When only one axis of d is set, that axis drives the scale. When d is zero,
// frame is returned unchanged.
func (d Dimension) Constrain(frame Dimension) Dimension {
	switch {
	case d.Complete():
		return frame.ScaleTo(d)
	case d.HasWidth():
		return frame.WithWidth(d.Width)
	case d.HasHeight():
		return frame.WithHeight(d.Height)
	default:
		return frame
	}
}
