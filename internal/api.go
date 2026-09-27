package internal

const (
	ToolShowWindow = "show_window"
	ToolHideWindow = "hide_window"
	ToolListPets   = "list_pets"
	ToolChangePet  = "change_pet"
	ToolPlayPrefix = "play_"
)

type ChangePetInput struct {
	Name      string `json:"name"`
	Animation string `json:"animation,omitempty"`
}

type ActiveEntry struct {
	Pet       string `json:"pet"`
	Animation string `json:"animation"`
}

type AnimEntry struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type PetEntry struct {
	Name        string      `json:"name"`
	Description *string     `json:"description"`
	Animations  []AnimEntry `json:"animations"`
}

type PetsResponse struct {
	Active ActiveEntry `json:"active"`
	Pets   []PetEntry  `json:"pets"`
}
