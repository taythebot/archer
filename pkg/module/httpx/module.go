package httpx

import "github.com/taythebot/archer/pkg/types"

// New creates a new Httpx client
func New(config types.HttpxConfig) *Module {
	return &Module{Config: config}
}

// Name of module
func (m *Module) Name() string {
	return "httpx"
}

func (m *Module) Payload() types.TaskPayload {
	return &Payload{}
}
