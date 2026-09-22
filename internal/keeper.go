package internal

import "sync"

type keeper struct {
	mu               *sync.RWMutex
	currentAnimation animation
	configFilePath   string
	pets             map[string][]animation
}

func NewKeeper(configFilePath string) (keeper, error) {
	k := keeper{
		configFilePath: configFilePath,
		mu:             &sync.RWMutex{},
	}
	if err := k.loadConfig(); err != nil {
		return keeper{}, err
	}
	return k, nil
}

func (k *keeper) loadConfig() error {
	cfg, err := LoadConfig(k.configFilePath)
	if err != nil {
		return err
	}

	pets := make(map[string][]animation, len(cfg.Pet))
	for _, pet := range cfg.Pet {
		anims := make([]animation, 0, len(pet.Animation))
		for _, a := range pet.Animation {
			fps := a.FPS
			if fps == 0 {
				fps = pet.FPS
			}
			windowHeight := a.Window.Height
			if windowHeight == 0 {
				windowHeight = pet.Window.Height
			}
			anims = append(anims, animation{
				filePath:     a.File,
				fps:          fps,
				height:       pet.Height,
				width:        pet.Width,
				name:         a.Name,
				description:  a.Description,
				windowHeight: windowHeight,
			})
		}
		pets[pet.Name] = anims
	}

	var current animation
	if anims, ok := pets[cfg.Server.Pet]; ok && len(anims) > 0 {
		current = anims[0]
	}

	k.mu.Lock()
	defer k.mu.Unlock()
	k.pets = pets
	k.currentAnimation = current
	return nil
}
