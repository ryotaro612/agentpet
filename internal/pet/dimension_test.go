package pet

import "testing"

func TestDimensionComplete(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		d    Dimension
		want bool
	}{
		{"reports true when both width and height are positive", Dimension{Width: 10, Height: 20}, true},
		{"reports false when width is zero", Dimension{Width: 0, Height: 20}, false},
		{"reports false when height is zero", Dimension{Width: 10, Height: 0}, false},
		{"reports false when both dimensions are zero", Dimension{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := c.d.Complete(); got != c.want {
				t.Errorf("Complete() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestDimensionHasWidth(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		d    Dimension
		want bool
	}{
		{"reports true when width is positive", Dimension{Width: 10}, true},
		{"reports false when width is zero", Dimension{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := c.d.HasWidth(); got != c.want {
				t.Errorf("HasWidth() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestDimensionHasHeight(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		d    Dimension
		want bool
	}{
		{"reports true when height is positive", Dimension{Height: 10}, true},
		{"reports false when height is zero", Dimension{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := c.d.HasHeight(); got != c.want {
				t.Errorf("HasHeight() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestDimensionScaleTo(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		d      Dimension
		target Dimension
		want   Dimension
	}{
		{
			"scales uniformly when source and target have the same aspect ratio",
			Dimension{Width: 100, Height: 100},
			Dimension{Width: 50, Height: 50},
			Dimension{Width: 50, Height: 50},
		},
		{
			"is constrained by width when the source is wider than the target aspect ratio",
			Dimension{Width: 200, Height: 100},
			Dimension{Width: 100, Height: 100},
			Dimension{Width: 100, Height: 50},
		},
		{
			"is constrained by height when the source is taller than the target aspect ratio",
			Dimension{Width: 100, Height: 200},
			Dimension{Width: 100, Height: 100},
			Dimension{Width: 50, Height: 100},
		},
		{
			"scales up when the target is larger than the source",
			Dimension{Width: 50, Height: 50},
			Dimension{Width: 100, Height: 100},
			Dimension{Width: 100, Height: 100},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := c.d.ScaleTo(c.target); got != c.want {
				t.Errorf("ScaleTo(%v) = %v, want %v", c.target, got, c.want)
			}
		})
	}
}

func TestDimensionWithWidth(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		d     Dimension
		width int
		want  Dimension
	}{
		{
			"infers height from aspect ratio when scaling down",
			Dimension{Width: 200, Height: 100},
			100,
			Dimension{Width: 100, Height: 50},
		},
		{
			"infers height from aspect ratio when scaling up",
			Dimension{Width: 100, Height: 50},
			200,
			Dimension{Width: 200, Height: 100},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := c.d.WithWidth(c.width); got != c.want {
				t.Errorf("WithWidth(%d) = %v, want %v", c.width, got, c.want)
			}
		})
	}
}

func TestDimensionWithHeight(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		d      Dimension
		height int
		want   Dimension
	}{
		{
			"infers width from aspect ratio when scaling down",
			Dimension{Width: 200, Height: 100},
			50,
			Dimension{Width: 100, Height: 50},
		},
		{
			"infers width from aspect ratio when scaling up",
			Dimension{Width: 100, Height: 50},
			100,
			Dimension{Width: 200, Height: 100},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := c.d.WithHeight(c.height); got != c.want {
				t.Errorf("WithHeight(%d) = %v, want %v", c.height, got, c.want)
			}
		})
	}
}
