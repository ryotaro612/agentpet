package internal

import (
	"flag"
	"log/slog"
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

func ParseArgs() (args, error) {
	var configFile string
	var verbose bool
	flag.StringVar(&configFile, "c", "", "path to config file (TOML)")
	flag.BoolVar(&verbose, "v", false, "enable verbose logging")
	flag.Parse()
	return args{ConfigFile: configFile, verbose: verbose}, nil
}
