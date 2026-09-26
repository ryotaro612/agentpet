package pet

import (
	"fmt"

	"github.com/ryotaro612/agentpet/internal/config"
)

type Keeper struct {
	pets             map[string][]Animation
	pet              string
	defaultAnimations map[string]string
}

func NewKeeper(cfg config.Config) Keeper {
	k := Keeper{
		pets:             buildPetsMap(cfg),
		defaultAnimations: make(map[string]string),
	}
	if _, ok := k.pets[cfg.Pet]; ok {
		k.pet = cfg.Pet
	} else {
		for name := range k.pets {
			k.pet = name
			break
		}
	}
	for petName, anims := range k.pets {
		if len(anims) > 0 {
			k.defaultAnimations[petName] = anims[0].Name
		}
	}
	return k
}

func (k Keeper) DefaultPet() string {
	return k.pet
}

func (k Keeper) DefaultAnim(petName string) (Animation, bool) {
	animName, ok := k.defaultAnimations[petName]
	if !ok {
		return Animation{}, false
	}
	for _, a := range k.pets[petName] {
		if a.Name == animName {
			return a, true
		}
	}
	return Animation{}, false
}

func (k Keeper) PlayAnimation(currentPet, name string) (Animation, error) {
	anims, ok := k.pets[currentPet]
	if !ok {
		return Animation{}, fmt.Errorf("current pet %q not found", currentPet)
	}
	for _, a := range anims {
		if a.Name != name {
			continue
		}
		if err := a.live(); err != nil {
			return Animation{}, fmt.Errorf("animation %q unavailable: %w", name, err)
		}
		return a, nil
	}
	return Animation{}, fmt.Errorf("animation %q not found for pet %q", name, currentPet)
}

// ChangePet switches to petName. If animName is non-empty the named animation
// is activated; otherwise the first live animation is used.
func (k Keeper) ChangePet(petName, animName string) (Animation, error) {
	anims, ok := k.pets[petName]
	if !ok {
		return Animation{}, fmt.Errorf("pet %q not found", petName)
	}
	if animName != "" {
		for _, a := range anims {
			if a.Name != animName {
				continue
			}
			if err := a.live(); err != nil {
				return Animation{}, fmt.Errorf("animation %q unavailable: %w", animName, err)
			}
			return a, nil
		}
		return Animation{}, fmt.Errorf("animation %q not found for pet %q", animName, petName)
	}
	for _, a := range anims {
		if a.live() != nil {
			continue
		}
		return a, nil
	}
	return Animation{}, fmt.Errorf("pet %q has no available animations", petName)
}

type PetInfo struct {
	Name       string
	Animations []Animation
}

func (k Keeper) AllPets() []PetInfo {
	infos := make([]PetInfo, 0, len(k.pets))
	for name, anims := range k.pets {
		infos = append(infos, PetInfo{Name: name, Animations: anims})
	}
	return infos
}

func (k Keeper) Animations(petName string) []Animation {
	return k.pets[petName]
}

func (k Keeper) OtherPetNames(currentPet string) []string {
	others := make([]string, 0, len(k.pets)-1)
	for name := range k.pets {
		if name != currentPet {
			others = append(others, name)
		}
	}
	return others
}

func buildPetsMap(cfg config.Config) map[string][]Animation {
	pets := make(map[string][]Animation, len(cfg.Pets))
	for _, p := range cfg.Pets {
		anims := make([]Animation, 0, len(p.Animations))
		for _, a := range p.Animations {
			anims = append(anims, resolveAnimation(a, p, cfg))
		}
		pets[p.Name] = anims
	}
	return pets
}
