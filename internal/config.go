package internal

import "github.com/BurntSushi/toml"

type Config struct {
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
}
