package masscan

import (
	"os"
	"strconv"
	"strings"
)

const baseConfig string = `randomize-hosts = true
interactive = true
exclude = 255.255.255.255
`

// addLine adds a key and value into the Masscan config file
func addLine(config, key, value string) string {
	return config + key + " = " + value + "\n"
}

// BuildConfig builds and creates a Masscan config file
func (m *Module) BuildConfig(taskId string, payload *Payload) (configFile string, err error) {
	// Copy base config
	config := baseConfig

	// Add node specific config
	if m.Config.Rate != 0 {
		config = addLine(config, "rate", strconv.Itoa(int(m.Config.Rate)))
	}
	if m.Config.AdapterPort != 0 {
		config = addLine(config, "adapter-port", strconv.Itoa(int(m.Config.AdapterPort)))
	}

	// Add targets and shards
	config = addLine(config, "range", strings.Join(payload.Targets, ","))
	config = addLine(config, "shard", payload.Shard)
	config = addLine(config, "seed", strconv.FormatInt(payload.Seed, 10))

	// Add ports
	var ports []string
	for _, port := range payload.Ports {
		ports = append(ports, strconv.FormatInt(int64(port), 10))
	}
	config = addLine(config, "ports", strings.Join(ports, ","))

	// Write to config directory
	configFile = m.Config.ConfigDir + "/masscan_" + taskId + ".conf"
	if err = os.WriteFile(configFile, []byte(config), 0644); err != nil {
		return configFile, err
	}

	return configFile, nil
}
