package dockerUtild

import (
	"fmt"
	"strings"
)

func ParseCreateContainerOptionFromYaml(serviceConfigFromYaml map[string]ServiceConfig, serviceName string) CreateContainerOption {
	serviceConfig := serviceConfigFromYaml[serviceName]
	// image name
	imageName := serviceConfig.Image
	// container name
	containerName := serviceConfig.ContainerName
	// container labels
	labels := serviceConfig.Labels
	// mounted volume
	mounts := make(map[string]string, 0)
	volumes := serviceConfig.Volumes
	for _, volume := range volumes {
		mounts[volume.Source] = volume.Target
	}
	// container environments
	enviromentsWithUpperKey := make(map[string]string)
	for envName, envValue := range serviceConfig.Environment {
		enviromentsWithUpperKey[strings.ToUpper(envName)] = envValue
	}
	environments := make([]string, 0)
	for envName, envValue := range enviromentsWithUpperKey {
		environments = append(environments, fmt.Sprintf("%s=%s", envName, envValue))
	}
	// container capabilities
	capabilities := serviceConfig.Capabilities
	// ports
	mappedPorts := make(map[string]string, 0)
	ports := serviceConfig.Ports
	for _, port := range ports {
		parts := strings.Split(port, ":")
		mappedPorts[parts[0]] = parts[1]
	}

	return CreateContainerOption{
		ImageName:     imageName,
		ContainerName: containerName,
		Labels:        labels,
		Mounts:        mounts,
		Environments:  environments,
		Capabilities:  capabilities,
		Ports:         mappedPorts,
	}
}
