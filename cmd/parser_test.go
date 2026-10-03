package main

import (
	"errors"
	"flag"
	"io"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		argv    []string
		want    parsedArgs
		wantErr string
		isHelp  bool
	}{
		{
			name: "parses port and a subcommand",
			argv: []string{"-p", "8080", "pets"},
			want: parsedArgs{port: 8080, command: "pets"},
		},
		{
			name: "parses the verbose flag",
			argv: []string{"-p", "8080", "-v", "pets"},
			want: parsedArgs{port: 8080, command: "pets", verbose: true},
		},
		{
			name: "parses an optional animation name for the animation subcommand",
			argv: []string{"-p", "8080", "animation", "idle"},
			want: parsedArgs{port: 8080, command: "animation", animation: animationArgs{animation: "idle"}},
		},
		{
			name: "parses the pet name and optional animation name for the pet subcommand",
			argv: []string{"-p", "8080", "pet", "cat", "idle"},
			want: parsedArgs{port: 8080, command: "pet", pet: petArgs{pet: "cat", animationArgs: animationArgs{animation: "idle"}}},
		},
		{
			name:    "returns an error when -p is not provided",
			argv:    []string{"pets"},
			wantErr: "-p <port> is required",
		},
		{
			name:    "returns an error when no subcommand is given",
			argv:    []string{"-p", "8080"},
			wantErr: "subcommand is required",
		},
		{
			name:    "returns an error for an unrecognised subcommand",
			argv:    []string{"-p", "8080", "unknown"},
			wantErr: `unknown subcommand "unknown"`,
		},
		{
			name: "selects a pet at random when no name is given",
			argv: []string{"-p", "8080", "pet"},
			want: parsedArgs{port: 8080, command: "pet"},
		},
		{
			name:   "prints help and returns ErrHelp when -h is passed",
			argv:   []string{"-h"},
			isHelp: true,
		},
		{
			name: "parses the vet subcommand",
			argv: []string{"-p", "8080", "vet"},
			want: parsedArgs{port: 8080, command: "vet"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := parse(c.argv, io.Discard)
			if c.isHelp {
				if !errors.Is(err, flag.ErrHelp) {
					t.Errorf("parse() error = %v, want flag.ErrHelp", err)
				}
				return
			}
			if c.wantErr != "" {
				if err == nil {
					t.Fatalf("parse(%v) = %+v, want error %q", c.argv, got, c.wantErr)
				}
				if err.Error() != c.wantErr {
					t.Errorf("parse(%v) error = %q, want %q", c.argv, err.Error(), c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse(%v) error = %v", c.argv, err)
			}
			if got != c.want {
				t.Errorf("parse(%v)\n  got  %+v\n  want %+v", c.argv, got, c.want)
			}
		})
	}
}
