package pet

type Dimension struct {
	Width  int
	Height int
}

// Complete reports whether both dimensions are set.
func (d Dimension) Complete() bool {
	return d.Width > 0 && d.Height > 0
}

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
