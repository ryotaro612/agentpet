package internal

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"
)

type args struct {
	ConfigFile string
	verbose    bool
}

func (a args) LogLevel() slog.Level {
	level := slog.LevelInfo
	if a.verbose {
		level = slog.LevelDebug
	}
	return level
}

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "agentpet", "config.toml"), nil
}

func ParseArgs() (args, error) {
	defaultConfig, err := defaultConfigPath()
	if err != nil {
		return args{}, err
	}
	var configFile string
	var verbose bool
	flag.StringVar(&configFile, "c", defaultConfig, "path to config file (TOML)")
	flag.BoolVar(&verbose, "v", false, "enable verbose logging")
	flag.Parse()
	return args{ConfigFile: configFile, verbose: verbose}, nil
}
