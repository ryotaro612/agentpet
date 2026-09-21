package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		toml    string
		want    Config
		wantErr bool
	}{
		{
			name: "parses server configuration fields",
			toml: `
[server]
port   = 8080
pet    = "cat"
height = 320
width  = 480
`,
			want: Config{
				Server: ServerConfig{Port: 8080, Pet: "cat", Height: 320, Width: 480},
			},
		},
		{
			name: "parses pet sprite dimensions and frame rate",
			toml: `
[pet.cat]
height = 32
width  = 32
fps    = 8
`,
			want: Config{
				Pet: map[string]PetConfig{
					"cat": {Height: 32, Width: 32, FPS: 8},
				},
			},
		},
		{
			name: "parses a pet animation with file path, name, and description",
			toml: `
[pet.cat.animation.idle]
file        = "idle.png"
fps         = 6
name        = "idle"
description = "The cat sits still"
`,
			want: Config{
				Pet: map[string]PetConfig{
					"cat": {
						Animation: map[string]AnimationConfig{
							"idle": {
								File:        "idle.png",
								FPS:         6,
								Name:        "idle",
								Description: "The cat sits still",
							},
						},
					},
				},
			},
		},
		{
			name: "applies window size overrides for a pet and its animation",
			toml: `
[pet.cat.window]
height = 320
width  = 480

[pet.cat.animation.walk.window]
height = 160
width  = 240
`,
			want: Config{
				Pet: map[string]PetConfig{
					"cat": {
						Window: &WindowConfig{Height: 320, Width: 480},
						Animation: map[string]AnimationConfig{
							"walk": {
								Window: &WindowConfig{Height: 160, Width: 240},
							},
						},
					},
				},
			},
		},
		{
			name:    "returns an error when the config file does not exist",
			toml:    "",
			want:    Config{},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var path string
			if c.wantErr {
				path = filepath.Join(t.TempDir(), "nonexistent.toml")
			} else {
				path = filepath.Join(t.TempDir(), "config.toml")
				if err := os.WriteFile(path, []byte(c.toml), 0600); err != nil {
					t.Fatal(err)
				}
			}

			got, err := LoadConfig(path)
			if (err != nil) != c.wantErr {
				t.Fatalf("LoadConfig() error = %v, wantErr %v", err, c.wantErr)
			}
			if !c.wantErr {
				compareConfigs(t, got, c.want)
			}
		})
	}
}

func compareConfigs(t *testing.T, got, want Config) {
	t.Helper()
	if got.Server != want.Server {
		t.Errorf("Server: got %+v, want %+v", got.Server, want.Server)
	}
	for name, wantPet := range want.Pet {
		gotPet, ok := got.Pet[name]
		if !ok {
			t.Errorf("pet %q not found in result", name)
			continue
		}
		if gotPet.Height != wantPet.Height || gotPet.Width != wantPet.Width || gotPet.FPS != wantPet.FPS {
			t.Errorf("pet %q: got %+v, want %+v", name, gotPet, wantPet)
		}
		compareWindows(t, "pet "+name, gotPet.Window, wantPet.Window)
		for animName, wantAnim := range wantPet.Animation {
			gotAnim, ok := gotPet.Animation[animName]
			if !ok {
				t.Errorf("pet %q animation %q not found", name, animName)
				continue
			}
			if gotAnim.File != wantAnim.File || gotAnim.FPS != wantAnim.FPS ||
				gotAnim.Name != wantAnim.Name || gotAnim.Description != wantAnim.Description {
				t.Errorf("pet %q animation %q: got %+v, want %+v", name, animName, gotAnim, wantAnim)
			}
			compareWindows(t, "pet "+name+" animation "+animName, gotAnim.Window, wantAnim.Window)
		}
	}
}

func compareWindows(t *testing.T, label string, got, want *WindowConfig) {
	t.Helper()
	if want == nil && got == nil {
		return
	}
	if (want == nil) != (got == nil) {
		t.Errorf("%s window: got %v, want %v", label, got, want)
		return
	}
	if *got != *want {
		t.Errorf("%s window: got %+v, want %+v", label, *got, *want)
	}
}
