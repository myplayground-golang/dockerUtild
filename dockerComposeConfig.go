package dockerUtild

type DockerYamlConfig struct {
	Services map[string]ServiceConfig `mapstructure:"services"`
	Networks NetworkConfig            `mapstructure:"networks"`
}

type VolumeBind struct {
	Type   string
	Source string
	Target string
}

type ServiceConfig struct {
	Image         string
	Ports         []string
	Environment   map[string]string
	Capabilities  []string `mapstructure:"cap_add"`
	Volumes       []VolumeBind
	ContainerName string `mapstructure:"container_name"`
	Command       []string
	Labels        map[string]string
	Networks      []string
}

type Network struct {
}

type NetworkConfig struct {
	Network Network
}
