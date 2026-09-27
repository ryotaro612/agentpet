package pet

import (
	"testing"

	"github.com/ryotaro612/agentpet/internal/config"
)

func TestKeeperDefaultPet(t *testing.T) {
	t.Parallel()

	t.Run("returns the explicitly configured default pet", func(t *testing.T) {
		t.Parallel()
		cfg := config.Config{
			Pet: "dog",
			Pets: []config.PetConfig{
				{Name: "cat", Frame: config.DimensionConfig{Width: 32, Height: 32},
					Animations: []config.AnimationConfig{{Name: "idle", File: "cat.png"}}},
				{Name: "dog", Frame: config.DimensionConfig{Width: 32, Height: 32},
					Animations: []config.AnimationConfig{{Name: "idle", File: "dog.png"}}},
			},
		}
		if got := NewKeeper(cfg).DefaultPet(); got != "dog" {
			t.Errorf("DefaultPet() = %q, want %q", got, "dog")
		}
	})

	t.Run("returns the only available pet when no default is configured", func(t *testing.T) {
		t.Parallel()
		cfg := config.Config{
			Pets: []config.PetConfig{
				{Name: "cat", Frame: config.DimensionConfig{Width: 32, Height: 32},
					Animations: []config.AnimationConfig{{Name: "idle", File: "cat.png"}}},
			},
		}
		if got := NewKeeper(cfg).DefaultPet(); got != "cat" {
			t.Errorf("DefaultPet() = %q, want %q", got, "cat")
		}
	})
}

func TestKeeperPlayAnimation(t *testing.T) {
	t.Parallel()

	t.Run("returns the animation when the pet and animation exist and the image is readable", func(t *testing.T) {
		t.Parallel()
		path := writePNG(t, 32, 32)
		cfg := config.Config{
			Pets: []config.PetConfig{{
				Name:  "cat",
				Frame: config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{
					{Name: "idle", File: path},
				},
			}},
		}
		anim, err := NewKeeper(cfg).PlayAnimation("cat", "idle")
		if err != nil {
			t.Fatalf("PlayAnimation() error = %v", err)
		}
		if anim.Name != "idle" {
			t.Errorf("PlayAnimation() name = %q, want %q", anim.Name, "idle")
		}
	})

	t.Run("returns an error when the animation name is not found for the pet", func(t *testing.T) {
		t.Parallel()
		path := writePNG(t, 32, 32)
		cfg := config.Config{
			Pets: []config.PetConfig{{
				Name:  "cat",
				Frame: config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{
					{Name: "idle", File: path},
				},
			}},
		}
		if _, err := NewKeeper(cfg).PlayAnimation("cat", "walk"); err == nil {
			t.Error("PlayAnimation() = nil, want error for unknown animation")
		}
	})

	t.Run("returns an error when the pet name is not found", func(t *testing.T) {
		t.Parallel()
		cfg := config.Config{
			Pets: []config.PetConfig{{
				Name:  "cat",
				Frame: config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{
					{Name: "idle", File: "idle.png"},
				},
			}},
		}
		if _, err := NewKeeper(cfg).PlayAnimation("dog", "idle"); err == nil {
			t.Error("PlayAnimation() = nil, want error for unknown pet")
		}
	})

	t.Run("returns an error when the image file cannot be read", func(t *testing.T) {
		t.Parallel()
		cfg := config.Config{
			Pets: []config.PetConfig{{
				Name:  "cat",
				Frame: config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{
					{Name: "idle", File: "/nonexistent/sprite.png"},
				},
			}},
		}
		if _, err := NewKeeper(cfg).PlayAnimation("cat", "idle"); err == nil {
			t.Error("PlayAnimation() = nil, want error for unreadable image")
		}
	})
}

func TestKeeperChangePet(t *testing.T) {
	t.Parallel()

	t.Run("returns the named animation when both pet and animation exist", func(t *testing.T) {
		t.Parallel()
		path := writePNG(t, 32, 32)
		cfg := config.Config{
			Pets: []config.PetConfig{{
				Name:  "dog",
				Frame: config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{
					{Name: "run", File: path},
				},
			}},
		}
		anim, err := NewKeeper(cfg).ChangePet("dog", "run")
		if err != nil {
			t.Fatalf("ChangePet() error = %v", err)
		}
		if anim.Name != "run" {
			t.Errorf("ChangePet() name = %q, want %q", anim.Name, "run")
		}
	})

	t.Run("returns the first live animation when no animation name is given", func(t *testing.T) {
		t.Parallel()
		livePath := writePNG(t, 32, 32)
		cfg := config.Config{
			Pets: []config.PetConfig{{
				Name:  "dog",
				Frame: config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{
					{Name: "dead", File: "/nonexistent/sprite.png"},
					{Name: "run", File: livePath},
				},
			}},
		}
		anim, err := NewKeeper(cfg).ChangePet("dog", "")
		if err != nil {
			t.Fatalf("ChangePet() error = %v", err)
		}
		if anim.Name != "run" {
			t.Errorf("ChangePet() name = %q, want %q", anim.Name, "run")
		}
	})

	t.Run("returns an error when the pet does not exist", func(t *testing.T) {
		t.Parallel()
		cfg := config.Config{
			Pets: []config.PetConfig{{
				Name:  "cat",
				Frame: config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{{Name: "idle", File: "idle.png"}},
			}},
		}
		if _, err := NewKeeper(cfg).ChangePet("dog", ""); err == nil {
			t.Error("ChangePet() = nil, want error for unknown pet")
		}
	})

	t.Run("returns an error when the specified animation does not exist for the pet", func(t *testing.T) {
		t.Parallel()
		path := writePNG(t, 32, 32)
		cfg := config.Config{
			Pets: []config.PetConfig{{
				Name:  "cat",
				Frame: config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{{Name: "idle", File: path}},
			}},
		}
		if _, err := NewKeeper(cfg).ChangePet("cat", "walk"); err == nil {
			t.Error("ChangePet() = nil, want error for unknown animation")
		}
	})

	t.Run("returns an error when no live animations are available for the pet", func(t *testing.T) {
		t.Parallel()
		cfg := config.Config{
			Pets: []config.PetConfig{{
				Name:  "cat",
				Frame: config.DimensionConfig{Width: 32, Height: 32},
				Animations: []config.AnimationConfig{
					{Name: "idle", File: "/nonexistent/sprite.png"},
				},
			}},
		}
		if _, err := NewKeeper(cfg).ChangePet("cat", ""); err == nil {
			t.Error("ChangePet() = nil, want error when all animations are unavailable")
		}
	})
}
