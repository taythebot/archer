package form

// NewScan is the validation struct for the "Create" scan controller
type NewScan struct {
	Targets     []string `json:"targets" validate:"required,min=1,unique"`
	Ports       []uint16 `json:"ports" validate:"required,min=1,unique"`
	Modules     []string `json:"modules" validate:"required,min=1,unique"`
	NucleiTypes []string `json:"nuclei_types,omitempty" validate:"omitempty,unique"`
	Arguments   string   `json:"arguments,omitempty"`
}
