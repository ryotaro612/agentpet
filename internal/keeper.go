package internal

import "github.com/ryotaro612/agentpet/internal/config"

func animationsFromConfig(cfg config.Config) []animation {
	var result []animation
	for _, pet := range cfg.Pets {
		for _, a := range pet.Animations {
			result = append(result, resolveAnimation(a, pet, cfg))
		}
	}
	return result
}

func resolveAnimation(a config.AnimationConfig, pet config.PetConfig, cfg config.Config) animation {
	return animation{
		filePath:     a.File,
		fps:          coalesce(a.FPS, pet.FPS, cfg.FPS),
		height:       coalesce(a.Frame.Height, pet.Frame.Height, cfg.Frame.Height),
		width:        coalesce(a.Frame.Width, pet.Frame.Width, cfg.Frame.Width),
		name:         a.Name,
		description:  a.Description,
		windowHeight: coalesce(a.Window.Height, pet.Window.Height, cfg.Window.Height),
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

func initialAnimation(cfg config.Config) animation {
	for _, pet := range cfg.Pets {
		if pet.Name == cfg.Pet && len(pet.Animations) > 0 {
			return resolveAnimation(pet.Animations[0], pet, cfg)
		}
	}
	return animation{}
}
