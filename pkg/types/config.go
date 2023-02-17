package types

type CoordinatorConfig struct {
	Listen        string            `yaml:"listen" validate:"required"`
	Postgresql    PostgresConfig    `yaml:"postgresql" validate:"required"`
	Redis         RedisClientConfig `yaml:"redis" validate:"required"`
	Elasticsearch ElasticConfig     `yaml:"elasticsearch" validate:"required"`
}

type WorkerConfig struct {
	Id            string            `yaml:"id" validate:"required,alphanum"`
	Concurrency   int               `yaml:"concurrency" validate:"gte=0"`
	Coordinator   string            `yaml:"coordinator" validate:"required,url"`
	Modules       []string          `yaml:"modules" validate:"required,min=1,unique"`
	Redis         RedisServerConfig `yaml:"redis" validate:"required"`
	Elasticsearch ElasticConfig     `yaml:"elasticsearch" validate:"required"`
	Masscan       *MasscanConfig    `yaml:"masscan,omitempty"`
	Httpx         *HttpxConfig      `yaml:"httpx,omitempty"`
	Nuclei        *NucleiConfig     `yaml:"nuclei,omitempty"`
}

type SchedulerConfig struct {
	Id            string            `yaml:"id" validate:"required,alphanum"`
	Concurrency   int               `yaml:"concurrency" validate:"gte=0"`
	Coordinator   string            `yaml:"coordinator" validate:"required,url"`
	Postgresql    PostgresConfig    `yaml:"postgresql" validate:"required"`
	RedisClient   RedisClientConfig `yaml:"redis_client" validate:"required"`
	RedisServer   RedisServerConfig `yaml:"redis_server" validate:"required"`
	Elasticsearch ElasticConfig     `yaml:"elasticsearch" validate:"required"`
}

type CliConfig struct {
	Coordinator string `yaml:"coordinator" validate:"required,url"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" validate:"required"`
	Port     int    `yaml:"port" validate:"required"`
	Database string `yaml:"database" validate:"required"`
	Username string `yaml:"username" validate:"required"`
	Password string `yaml:"password" validate:"required"`
}

type RedisClientConfig struct {
	Host          string `yaml:"host" validate:"required,hostname_port"`
	Username      string `yaml:"username,omitempty"`
	Password      string `yaml:"password,omitempty"`
	Database      int    `yaml:"database" validate:"gte=0"`
	TaskRetention string `yaml:"task_retention,omitempty" validate:"duration"`
}

type RedisServerConfig struct {
	Host      string `yaml:"host" validate:"required,hostname_port"`
	Username  string `yaml:"username,omitempty"`
	Password  string `yaml:"password,omitempty"`
	Database  int    `yaml:"database" validate:"gte=0"`
	Heartbeat int    `yaml:"heartbeat" validate:"gt=0"`
}

type ElasticConfig struct {
	Hosts    []string `yaml:"hosts" validate:"required,min=1,unique"`
	Username string   `yaml:"username,omitempty"`
	Password string   `yaml:"password,omitempty"`
	Index    string   `yaml:"index" validate:"required"`
	Bulk     struct {
		FlushBytes    int    `yaml:"flush_bytes"`
		FlushInterval int    `yaml:"flush_interval"`
		Pipeline      string `yaml:"pipeline,omitempty"`
	} `yaml:"bulk,omitempty"`
}
