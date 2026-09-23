package config

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

func LoadConfig(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	var errs []error
	if c.Port < 0 || c.Port > 65535 {
		errs = append(errs, fmt.Errorf("port %d is out of range: must be between 0 and 65535", c.Port))
	}
	if c.Frame.isPartial() {
		errs = append(errs, fmt.Errorf("config frame: both height and width are required"))
	}
	petSeen := make(map[string]bool, len(c.Pets))
	for _, pet := range c.Pets {
		if petSeen[pet.Name] {
			errs = append(errs, fmt.Errorf("duplicate pet name %q", pet.Name))
		}
		petSeen[pet.Name] = true
		if pet.Frame.isPartial() {
			errs = append(errs, fmt.Errorf("pet %q frame: both height and width are required", pet.Name))
		}
		if len(pet.Animations) == 0 {
			errs = append(errs, fmt.Errorf("pet %q has no animations", pet.Name))
		}
		animSeen := make(map[string]bool, len(pet.Animations))
		for _, anim := range pet.Animations {
			if animSeen[anim.Name] {
				errs = append(errs, fmt.Errorf("pet %q has duplicate animation name %q", pet.Name, anim.Name))
			}
			animSeen[anim.Name] = true
			if anim.Frame.isPartial() {
				errs = append(errs, fmt.Errorf("animation %q frame: both height and width are required", anim.Name))
			}
			if !anim.Frame.isComplete() && !pet.Frame.isComplete() && !c.Frame.isComplete() {
				errs = append(errs, fmt.Errorf("animation %q in pet %q: frame is required", anim.Name, pet.Name))
			}
			if anim.File == "" {
				errs = append(errs, fmt.Errorf("animation %q in pet %q: filepath is required", anim.Name, pet.Name))
			}
		}
	}
	if c.Pet != "" {
		found := false
		for _, pet := range c.Pets {
			if pet.Name == c.Pet {
				found = true
				break
			}
		}
		if !found {
			errs = append(errs, fmt.Errorf("pet %q not found in pets", c.Pet))
		}
	}
	return errors.Join(errs...)
}
