package pet

import (
	"fmt"

	"github.com/ryotaro612/agentpet/internal/config"
)

type Keeper struct {
	pets       map[string][]Animation
	currentPet string
	current    Animation
}

func NewKeeper(cfg config.Config) Keeper {
	k := Keeper{pets: buildPetsMap(cfg)}
	if _, ok := k.pets[cfg.Pet]; ok {
		k.currentPet = cfg.Pet
	} else {
		for name := range k.pets {
			k.currentPet = name
			break
		}
	}
	if anims := k.pets[k.currentPet]; len(anims) > 0 {
		k.current = anims[0]
	}
	return k
}

func (k Keeper) PlayAnimation(name string) (Keeper, error) {
	anims, ok := k.pets[k.currentPet]
	if !ok {
		return k, fmt.Errorf("current pet %q not found", k.currentPet)
	}
	for _, a := range anims {
		if a.Name != name {
			continue
		}
		if err := a.live(); err != nil {
			return k, fmt.Errorf("animation %q unavailable: %w", name, err)
		}
		k.current = a
		return k, nil
	}
	return k, fmt.Errorf("animation %q not found for pet %q", name, k.currentPet)
}

// ChangePet switches to petName. If animName is non-empty the named animation
// is activated; otherwise the first live animation is used.
func (k Keeper) ChangePet(petName, animName string) (Keeper, error) {
	anims, ok := k.pets[petName]
	if !ok {
		return k, fmt.Errorf("pet %q not found", petName)
	}
	if animName != "" {
		for _, a := range anims {
			if a.Name != animName {
				continue
			}
			if err := a.live(); err != nil {
				return k, fmt.Errorf("animation %q unavailable: %w", animName, err)
			}
			k.currentPet = petName
			k.current = a
			return k, nil
		}
		return k, fmt.Errorf("animation %q not found for pet %q", animName, petName)
	}
	for _, a := range anims {
		if a.live() != nil {
			continue
		}
		k.currentPet = petName
		k.current = a
		return k, nil
	}
	return k, fmt.Errorf("pet %q has no available animations", petName)
}

func (k Keeper) CurrentAnimation() Animation {
	return k.current
}

func (k Keeper) CurrentPetName() string {
	return k.currentPet
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

func (k Keeper) Animations() []Animation {
	return k.pets[k.currentPet]
}

func (k Keeper) OtherPetNames() []string {
	others := make([]string, 0, len(k.pets)-1)
	for name := range k.pets {
		if name != k.currentPet {
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
