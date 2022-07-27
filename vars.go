package dockerUtild

import (
	"github.com/spf13/viper"
)

var DockerConfigYamlReader *viper.Viper
var YamlServicesMap map[string]ServiceConfig
