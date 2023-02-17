package nuclei

import (
	"os"
	"strings"
)

// BuildConfig builds and creates a Nuclei targets file
func (m *Module) BuildConfig(taskId string, payload *Payload) (configFile string, err error) {
	// Write to config directory
	configFile = m.Config.ConfigDir + "/nuclei_" + taskId + "_targets.txt"
	if err = os.WriteFile(configFile, []byte(strings.Join(payload.Targets, "\n")), 0644); err != nil {
		return configFile, err
	}

	return configFile, nil
}
