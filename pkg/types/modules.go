package types

// Modules for scan
var Modules = []string{"masscan", "httpx", "nuclei"}

// Stages for modules
var Stages = []Stage{
	{
		Name: "masscan",
		Next: []string{"httpx", "nuclei"},
	},
	{
		Name:     "httpx",
		Previous: []string{"masscan"},
		Next:     []string{"nuclei"},
	},
	{
		Name:     "nuclei",
		Previous: []string{"masscan", "httpx"},
	},
}

// Stage for module
type Stage struct {
	Name     string   // Name of module
	Previous []string // Previous stages
	Next     []string // Next stages
}

type MasscanConfig struct {
	// Binary file path for Masscan
	Binary string `yaml:"binary" valdiate:"required"`
	// ConfigDir contains the directory to create Masscan configs in
	ConfigDir string `yaml:"config_dir" validate:"required,dir"`
	// ExcludeFile for Masscan
	ExcludeFile string `yaml:"exclude_file,omitempty" validate:"omitempty,file"`
	// Rate for Masscan
	Rate uint `yaml:"rate,omitempty" validate:"gte=1,omitempty"`
	// AdapterPort for Masscan
	AdapterPort uint16 `yaml:"adapter_port,omitempty"`
	// PersistConfig will persist Masscan configs after completion
	PersistConfig bool `yaml:"persist_config"`
}

type HttpxConfig struct {
	// Binary file path for Httpx
	Binary string `yaml:"binary" validate:"required"`
	// ConfigDir contains the directory to create Httpx target files in
	ConfigDir string `yaml:"config_dir" validate:"required,dir"`
	// HttpProxy is the value for the Httpx flag "-http-proxy"
	HttpProxy string `yaml:"http_proxy,omitempty" validate:"omitempty,url"`
	// SocksProxy is the value for the Httpx flag "-socks-proxy"
	SocksProxy string `yaml:"socks_proxy,omitempty"`
	// Threads is the value for the Httpx flag "-threads"
	Threads uint `yaml:"threads,omitempty" validate:"omitempty,gte=1"`
	// RateLimit is the value for the Httpx flag "-rl"
	RateLimit uint `yaml:"rate_limit,omitempty" validate:"omitempty,gte=1"`
	// PersistConfig will persist Httpx configs after completion
	PersistConfig bool `yaml:"persist_config"`
}

type NucleiConfig struct {
	// Binary file path for Nuclei
	Binary string `yaml:"binary" validate:"required"`
	// ConfigDir contains the directory to create Nuclei target files in
	ConfigDir string `yaml:"config_dir" validate:"required,dir"`
	// Proxies is the value for the Nuclei flag "-proxy"
	Proxies []string `yaml:"proxies,omitempty" validate:"omitempty,unique"`
	// Timeout is the value for the Nuclei flag "-timeout"
	Timeout uint `yaml:"timeout,omitempty" validate:"omitempty,gte=1"`
	// Retries is the value for the Nuclei flag "-retries", default is 1
	Retries uint `yaml:"retries,omitempty" validate:"omitempty,gte=1"`
	// RateLimit is the value for the Nuclei flag "-rl", default is 150
	RateLimit uint `yaml:"rate_limit,omitempty" validate:"omitempty,gte=1"`
	// BulkSize is the value for the Nuclei flag "-bs", default is 200
	BulkSize uint `yaml:"bulk_size,omitempty" validate:"omitempty,gte=1"`
	// Concurrency is the value for the Nuclei flag "-c", default is 50
	Concurrency uint `yaml:"concurrency,omitempty" validate:"omitempty,gte=1"`
	// PersistConfig will persist Nuclei configs after completion
	PersistConfig bool `yaml:"persist_config"`
}
