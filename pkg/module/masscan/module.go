package masscan

import "github.com/taythebot/archer/pkg/types"

// New creates a new Masscan client
func New(config types.MasscanConfig) *Module {
	return &Module{Config: config}
}

// Name of module
func (m *Module) Name() string {
	return "masscan"
}

func (m *Module) Payload() types.TaskPayload {
	return &Payload{}
}
