package internal

import (
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgsUsesConfigPathFromEnvironment(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "custom.toml")
	t.Setenv("AGENT_PET_CONFIG", configPath)

	previousArgs := os.Args
	previousFlags := flag.CommandLine
	t.Cleanup(func() {
		os.Args = previousArgs
		flag.CommandLine = previousFlags
	})
	os.Args = []string{"agentpet"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)

	got, err := ParseArgs()
	if err != nil {
		t.Fatalf("ParseArgs() error = %v", err)
	}
	if got.ConfigFile != configPath {
		t.Errorf("ConfigFile = %q, want %q", got.ConfigFile, configPath)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	t.Parallel()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		want string
	}{
		{"returns a path rooted at the user home directory", filepath.Join(home, ".local", "agentpet", "config.toml")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := defaultConfigPath()
			if err != nil {
				t.Fatalf("defaultConfigPath() error = %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestLogLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		verbose bool
		want    slog.Level
	}{
		{"returns debug level when verbose is enabled", true, slog.LevelDebug},
		{"returns info level when verbose is disabled", false, slog.LevelInfo},
	}
	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			a := args{verbose: c.verbose}
			if got := a.LogLevel(); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}
