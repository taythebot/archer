package nuclei

import "github.com/taythebot/archer/pkg/types"

// New creates a new Nuclei client
func New(config types.NucleiConfig) *Module {
	return &Module{Config: config}
}

// Name of module
func (m *Module) Name() string {
	return "nuclei"
}

func (m *Module) Payload() types.TaskPayload {
	return &Payload{}
}
