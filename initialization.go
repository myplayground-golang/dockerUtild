package dockerUtild

import (
	"fmt"
	"github.com/spf13/viper"
)

func Initialize() error {
	var err error

	// viper initial
	DockerConfigYamlReader = viper.New()
	//fmt.Println(GetCurrentWorkingFolder())

	DockerConfigYamlReader.AddConfigPath(GetCurrentWorkingFolder())
	DockerConfigYamlReader.SetConfigName("docker-compose")
	DockerConfigYamlReader.SetConfigType("yaml")

	if err = DockerConfigYamlReader.ReadInConfig(); err != nil {
		panic(err)
	} else {
		fmt.Println("Viper read docker-compose.yaml successfully")
	}

	YamlServicesMap = getServiceConfigFromYaml(DockerConfigYamlReader)
	//fmt.Println(YamlServicesMap)

	return nil
}
