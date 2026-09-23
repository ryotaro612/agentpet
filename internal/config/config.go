package config

func (d DimensionConfig) isPartial() bool {
	return (d.Height == 0) != (d.Width == 0)
}

func (d DimensionConfig) isComplete() bool {
	return d.Height > 0 && d.Width > 0
}

type Config struct {
	Port   int             `toml:"port"`
	Pet    string          `toml:"pet"`
	Window DimensionConfig `toml:"window"`
	Frame  DimensionConfig `toml:"frame"`
	FPS    int             `toml:"fps"`
	Pets   []PetConfig     `toml:"pets"`
}

type PetConfig struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description"`
	Frame       DimensionConfig   `toml:"frame"`
	Window      DimensionConfig   `toml:"window"`
	FPS         int               `toml:"fps"`
	Animations  []AnimationConfig `toml:"animations"`
}

type AnimationConfig struct {
	Name        string          `toml:"name"`
	Description string          `toml:"description"`
	File        string          `toml:"filepath"`
	Frame       DimensionConfig `toml:"frame"`
	Window      DimensionConfig `toml:"window"`
	FPS         int             `toml:"fps"`
}

type DimensionConfig struct {
	Height int `toml:"height"`
	Width  int `toml:"width"`
}
