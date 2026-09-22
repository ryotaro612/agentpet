package internal

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
			file: "testdata/server_config_fields.toml",
			want: Config{
				Server: ServerConfig{Port: 8080, Pet: "cat", Height: 320, Width: 480},
			},
		},
		{
			name: "parses pet sprite dimensions and frame rate",
			file: "testdata/pet_sprite.toml",
			want: Config{
				Pet: []PetConfig{
					{Name: "cat", Height: 32, Width: 32, FPS: 8},
				},
			},
		},
		{
			name: "parses a pet animation with file path, name, and description",
			file: "testdata/pet_animation.toml",
			want: Config{
				Pet: []PetConfig{
					{
						Name: "cat",
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
				Pet: []PetConfig{
					{
						Name:   "cat",
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
				if diff := cmp.Diff(c.want, got); diff != "" {
					t.Errorf("LoadConfig() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
