package runner

import (
	"github.com/taythebot/archer/pkg/types"
)

// defaultConfig for worker
func defaultConfig() *types.WorkerConfig {
	return &types.WorkerConfig{
		Nuclei: &types.NucleiConfig{
			Timeout:     5,
			Retries:     1,
			BulkSize:    200,
			Concurrency: 500,
		},
	}
}
