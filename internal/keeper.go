package internal

type keeper struct {
	currentAnimation animation
	configFilePath   string
	pets             map[string][]animation
	serverPort       int
}

func (k keeper) Port() int {
	return k.serverPort
}

func NewKeeper(configFilePath string) (keeper, error) {
	k := keeper{
		configFilePath: configFilePath,
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
	k.pets = pets
	k.serverPort = cfg.Server.Port

	if anims, ok := pets[cfg.Server.Pet]; ok && len(anims) > 0 {
		k.currentAnimation = anims[0]
	}
	return nil
}
