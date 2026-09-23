package internal

import (
	"fmt"

	"github.com/ryotaro612/agentpet/internal/config"
)

type keeper struct {
	pets       map[string][]animation
	currentPet string
	current    *animation
}

func newKeeper(cfg config.Config) keeper {
	k := keeper{pets: buildPetsMap(cfg)}
	if _, ok := k.pets[cfg.Pet]; ok {
		k.currentPet = cfg.Pet
	} else {
		for name := range k.pets {
			k.currentPet = name
			break
		}
	}
	if anims := k.pets[k.currentPet]; len(anims) > 0 {
		a := anims[0]
		k.current = &a
	}
	return k
}

func (k *keeper) playAnimation(name string) (*animation, error) {
	anims, ok := k.pets[k.currentPet]
	if !ok {
		return nil, fmt.Errorf("current pet %q not found", k.currentPet)
	}
	for _, a := range anims {
		if a.name != name {
			continue
		}
		if err := a.live(); err != nil {
			return nil, fmt.Errorf("animation %q unavailable: %w", name, err)
		}
		k.current = &a
		return k.current, nil
	}
	return nil, fmt.Errorf("animation %q not found for pet %q", name, k.currentPet)
}

func (k *keeper) changePet(name string) (*animation, error) {
	anims, ok := k.pets[name]
	if !ok {
		return nil, fmt.Errorf("pet %q not found", name)
	}
	for _, a := range anims {
		if a.live() != nil {
			continue
		}
		k.currentPet = name
		k.current = &a
		return k.current, nil
	}
	return nil, fmt.Errorf("pet %q has no available animations", name)
}

func (k *keeper) currentAnimation() *animation {
	return k.current
}

func (k *keeper) animationNames() []string {
	anims := k.pets[k.currentPet]
	names := make([]string, len(anims))
	for i, a := range anims {
		names[i] = a.name
	}
	return names
}

func (k *keeper) otherPetNames() []string {
	others := make([]string, 0, len(k.pets)-1)
	for name := range k.pets {
		if name != k.currentPet {
			others = append(others, name)
		}
	}
	return others
}

func buildPetsMap(cfg config.Config) map[string][]animation {
	pets := make(map[string][]animation, len(cfg.Pets))
	for _, pet := range cfg.Pets {
		anims := make([]animation, 0, len(pet.Animations))
		for _, a := range pet.Animations {
			anims = append(anims, resolveAnimation(a, pet, cfg))
		}
		pets[pet.Name] = anims
	}
	return pets
}

func resolveAnimation(a config.AnimationConfig, pet config.PetConfig, cfg config.Config) animation {
	return animation{
		filePath: a.File,
		fps:      coalesce(a.FPS, pet.FPS, cfg.FPS),
		frame: dimension{
			height: coalesce(a.Frame.Height, pet.Frame.Height, cfg.Frame.Height),
			width:  coalesce(a.Frame.Width, pet.Frame.Width, cfg.Frame.Width),
		},
		name:        a.Name,
		description: a.Description,
		window: dimension{
			height: coalesce(a.Window.Height, pet.Window.Height, cfg.Window.Height),
			width:  coalesce(a.Window.Width, pet.Window.Width, cfg.Window.Width),
		},
	}
}

func coalesce(vals ...int) int {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}
