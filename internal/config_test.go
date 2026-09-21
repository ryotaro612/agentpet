package internal

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		file    string
		want    Config
		wantErr bool
	}{
		{
			name: "parses server configuration fields",
			file: "testdata/server.toml",
			want: Config{
				Server: ServerConfig{Port: 8080, Pet: "cat", Height: 320, Width: 480},
			},
		},
		{
			name: "parses pet sprite dimensions and frame rate",
			file: "testdata/pet_sprite.toml",
			want: Config{
				Pet: map[string]PetConfig{
					"cat": {Height: 32, Width: 32, FPS: 8},
				},
			},
		},
		{
			name: "parses a pet animation with file path, name, and description",
			file: "testdata/pet_animation.toml",
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
			file: "testdata/window_overrides.toml",
			want: Config{
				Pet: map[string]PetConfig{
					"cat": {
						Window: WindowConfig{Height: 320, Width: 480},
						Animation: map[string]AnimationConfig{
							"walk": {
								Window: WindowConfig{Height: 160, Width: 240},
							},
						},
					},
				},
			},
		},
		{
			name:    "returns an error when the config file does not exist",
			file:    "testdata/nonexistent.toml",
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got, err := LoadConfig(c.file)
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
		if gotPet.Window != wantPet.Window {
			t.Errorf("pet %q window: got %+v, want %+v", name, gotPet.Window, wantPet.Window)
		}
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
			if gotAnim.Window != wantAnim.Window {
				t.Errorf("pet %q animation %q window: got %+v, want %+v", name, animName, gotAnim.Window, wantAnim.Window)
			}
		}
	}
}
