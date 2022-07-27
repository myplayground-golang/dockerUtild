package dockerUtild

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func getServiceConfigFromYaml(dockerConfigYamlReader *viper.Viper) map[string]ServiceConfig {
	var cfg DockerYamlConfig
	err := dockerConfigYamlReader.Unmarshal(&cfg)
	if err != nil {
		panic(fmt.Errorf("Unable to decode Config: %s \n", err))
	}
	return cfg.Services
}

func getContainerConfig(image string, exposedPorts nat.PortSet, labels map[string]string, environment []string, cmd []string) *container.Config {
	config := &container.Config{}
	config.Image = image
	config.ExposedPorts = exposedPorts
	config.Labels = labels
	config.Env = environment
	config.Cmd = cmd
	return config
}

func getExposedPortMap(portPairs map[string]string) (nat.PortSet, error) {
	// a map, key is host port, value is nat.PortSet(map)
	exposedContainerPorts := make(nat.PortSet, 0)

	for _, containerPort := range portPairs {
		cPort, _ := nat.NewPort("tcp", containerPort)
		exposedContainerPorts[cPort] = struct{}{}
	}

	return exposedContainerPorts, nil
}

func getPortBindings(mapping map[string]string, ports *nat.PortSet) (nat.PortMap, error) {
	portMap := make(nat.PortMap, 0)

	for port, _ := range *ports {
		// get host port from mapping
		targtetContainerPort := port.Port()
		targetHostPort := "-1"
		for hostP, containerP := range mapping {
			if containerP == targtetContainerPort {
				targetHostPort = hostP
			}
		}
		if targetHostPort == "-1" {
			return portMap, errors.New("Unable to find mapped host port for " + targetHostPort)
		}

		portBinding := nat.PortBinding{HostPort: targetHostPort}
		portBindingSlice := make([]nat.PortBinding, 0)
		portBindingSlice = append(portBindingSlice, portBinding)

		portMap[port] = portBindingSlice
	}
	return portMap, nil
}

func getMounts(bindPaths map[string]string) []mount.Mount {
	mounts := make([]mount.Mount, 0)
	for source, target := range bindPaths {
		m := mount.Mount{
			Type:   mount.TypeBind,
			Source: source,
			Target: target,
		}
		mounts = append(mounts, m)
	}
	return mounts
}

func getContainerHostConfig(portMap *nat.PortMap, mounts *[]mount.Mount, capAdd strslice.StrSlice) *container.HostConfig {
	return &container.HostConfig{
		PortBindings: *portMap,
		Mounts:       *mounts,
		CapAdd:       capAdd,
	}
}
