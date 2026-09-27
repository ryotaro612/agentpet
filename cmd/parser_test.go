package main

import (
	"errors"
	"flag"
	"io"
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		argv    []string
		want    args
		wantErr string
		isHelp  bool
	}{
		{
			name: "parses port and a subcommand",
			argv: []string{"-p", "8080", "pets"},
			want: args{Port: 8080, Command: "pets"},
		},
		{
			name: "parses the verbose flag",
			argv: []string{"-p", "8080", "-v", "pets"},
			want: args{Port: 8080, Verbose: true, Command: "pets"},
		},
		{
			name: "treats arguments after the subcommand as rest",
			argv: []string{"-p", "8080", "animation", "idle"},
			want: args{Port: 8080, Command: "animation", Rest: []string{"idle"}},
		},
		{
			name: "treats the pet name and optional animation as rest",
			argv: []string{"-p", "8080", "pet", "cat", "idle"},
			want: args{Port: 8080, Command: "pet", Rest: []string{"cat", "idle"}},
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
			name:   "prints help and returns ErrHelp when -h is passed",
			argv:   []string{"-h"},
			isHelp: true,
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
			if got.Port != c.want.Port {
				t.Errorf("Port = %d, want %d", got.Port, c.want.Port)
			}
			if got.Verbose != c.want.Verbose {
				t.Errorf("Verbose = %v, want %v", got.Verbose, c.want.Verbose)
			}
			if got.Command != c.want.Command {
				t.Errorf("Command = %q, want %q", got.Command, c.want.Command)
			}
			if !slices.Equal(got.Rest, c.want.Rest) {
				t.Errorf("Rest = %v, want %v", got.Rest, c.want.Rest)
			}
		})
	}
}
