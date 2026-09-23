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
				Port:   8080,
				Window: ResolutionConfig{Height: 320, Width: 480},
			},
		},
		{
			name: "parses pet sprite dimensions and frame rate",
			file: "testdata/pet_sprite.toml",
			want: Config{
				Pets: []PetConfig{
					{
						Name:  "cat",
						FPS:   8,
						Frame: ResolutionConfig{Height: 32, Width: 32},
						Animations: []AnimationConfig{
							{Name: "idle", File: "idle.png"},
						},
					},
				},
			},
		},
		{
			name: "parses a pet animation with file path, name, and description",
			file: "testdata/pet_animation.toml",
			want: Config{
				Pets: []PetConfig{
					{
						Name:  "cat",
						Frame: ResolutionConfig{Height: 32, Width: 32},
						Animations: []AnimationConfig{
							{
								Name:        "idle",
								Description: "The cat sits still",
								File:        "idle.png",
								FPS:         6,
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
				Pets: []PetConfig{
					{
						Name:   "cat",
						Frame:  ResolutionConfig{Height: 32, Width: 32},
						Window: ResolutionConfig{Height: 320, Width: 480},
						Animations: []AnimationConfig{
							{
								Name:   "walk",
								File:   "walk.png",
								Window: ResolutionConfig{Height: 160, Width: 240},
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
		{
			name:    "returns an error if the port is outside the valid range",
			file:    "testdata/config/invalid_port_negative.toml",
			wantErr: true,
		},
		{
			name: "accepts a port within the valid range",
			file: "testdata/config/valid_port.toml",
		},
		{
			name:    "returns an error if frame is missing",
			file:    "testdata/config/frame_missing.toml",
			wantErr: true,
		},
		{
			name:    "returns an error if frame defines only height",
			file:    "testdata/config/frame_height_only.toml",
			wantErr: true,
		},
		{
			name:    "returns an error if frame defines only width",
			file:    "testdata/config/frame_width_only.toml",
			wantErr: true,
		},
		{
			name: "accepts an animation that inherits frame from its pet",
			file: "testdata/config/frame_inherited_from_pet.toml",
		},
		{
			name: "accepts an animation that inherits frame from the config",
			file: "testdata/config/frame_inherited_from_config.toml",
		},
		{
			name:    "returns an error if filepath is missing",
			file:    "testdata/config/file_missing.toml",
			wantErr: true,
		},
		{
			name:    "returns an error if Config.Pet is not in Config.Pets",
			file:    "testdata/config/pet_not_found.toml",
			wantErr: true,
		},
		{
			name: "accepts Config.Pet that matches a pet name",
			file: "testdata/config/pet_found.toml",
		},
		{
			name:    "returns an error if a pet has no animations",
			file:    "testdata/config/animations_missing.toml",
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
			if !c.wantErr && !cmp.Equal(c.want, Config{}) {
				if diff := cmp.Diff(c.want, got); diff != "" {
					t.Errorf("LoadConfig() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
