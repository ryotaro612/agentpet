package internal

import "github.com/BurntSushi/toml"

type Config struct {
	Server ServerConfig         `toml:"server"`
	Pet    map[string]PetConfig `toml:"pet"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
}

type ServerConfig struct {
	Port   int    `toml:"port"`
	Pet    string `toml:"pet"`
	Height int    `toml:"height"`
	Width  int    `toml:"width"`
}

type PetConfig struct {
	Height    int                        `toml:"height"`
	Width     int                        `toml:"width"`
	FPS       int                        `toml:"fps"`
	Window    WindowConfig               `toml:"window"`
	Animation map[string]AnimationConfig `toml:"animation"`
}

type AnimationConfig struct {
	File        string       `toml:"file"`
	FPS         int          `toml:"fps"`
	Name        string       `toml:"name"`
	Description string       `toml:"description"`
	Window      WindowConfig `toml:"window"`
}

type WindowConfig struct {
	Height int `toml:"height"`
	Width  int `toml:"width"`
}
