package dockerUtild

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var SameNameContainerError = errors.New("multiple container with same container name found")
var NoImageFoundError = errors.New("no docker image found")
var SearchMultipleImageError = errors.New("multiple docker image found")

type DockerUtild struct {
	client *client.Client
}

func getDockerClient() *client.Client {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return dockerClient
}

/*

	public constructor

*/
func NewDockerUtild() DockerUtild {
	client := getDockerClient()
	return DockerUtild{
		client: client,
	}
}

/*
	common function parts
*/
func (d *DockerUtild) FetchEmbedClient() *client.Client {
	return d.client
}

func (d *DockerUtild) GetShortId(longId string) string {
	if strings.HasPrefix(longId, "sha256:") {
		return longId[7:19]
	} else {
		return longId[0:12]
	}
}

/*

	container related operation functions

*/
type CreateContainerOption struct {
	ImageName     string
	ContainerName string
	Labels        map[string]string
	Mounts        map[string]string
	Environments  []string
	Capabilities  []string
	Ports         map[string]string
	Commands      []string
}

func (d *DockerUtild) CreateContainerWithOption(ctx context.Context, option CreateContainerOption) (string, error) {
	return d.CreateContainer(
		ctx,
		option.ImageName,
		option.ContainerName,
		option.Labels,
		option.Mounts,
		option.Environments,
		option.Capabilities,
		option.Ports,
		option.Commands,
	)
}

func (d *DockerUtild) CreateContainer(
	ctx context.Context,
	image string, containerName string,
	labels map[string]string,
	mounts map[string]string,
	environments []string,
	capabilities []string,
	mappedPorts map[string]string,
	commands []string) (string, error) {

	// See:
	// https://docs.docker.com/engine/api/v1.41/#operation/ContainerCreate

	// create container.Config
	exposedPorts, err := getExposedPortMap(mappedPorts)
	if err != nil {
		return "", err
	}
	config := getContainerConfig(image, exposedPorts, labels, environments, commands)

	// create container.HostConfig
	portMap, _ := getPortBindings(mappedPorts, &exposedPorts)
	mountsConfig := getMounts(mounts)
	hostConfig := getContainerHostConfig(&portMap, &mountsConfig, capabilities)

	// create network.NetworkingConfig
	//networkingConfig := &network.NetworkingConfig{}

	// create  *specs.Platform{}
	//platform := &specs.Platform{}

	// containerName
	createContainerName := containerName

	body, err := d.client.ContainerCreate(ctx, config, hostConfig, nil, nil, createContainerName)
	if err != nil {
		return "", err
	}

	return body.ID, nil
}

func (d *DockerUtild) ClearAllImageContainer(ctx context.Context, imageName string, imageTag string, skipRunning bool) {
	//clearedContainerIDs := make([]string, 0)
	//failToClearContainerIDs := make([]string, 0)

	// See:
	// https://docs.docker.com/engine/api/v1.41/#operation/ContainerList
	containerFilter := filters.NewArgs(filters.Arg("ancestor", fmt.Sprintf("%s:%s", imageName, imageTag)))
	containerListOptions := types.ContainerListOptions{
		All:     true,
		Filters: containerFilter,
	}
	containers, err := d.client.ContainerList(ctx, containerListOptions)
	if err != nil {
		log.Printf("ClearAllImageContainer %s", err.Error())
		return
	}

	for _, container := range containers {
		log.Printf("container id %s, state %s\n", container.ID[0:12], container.State)
		if skipRunning {
			if container.State == "running" {
				continue
			}
		}
		go d.StopAndRemoveContainer(ctx, container.ID)
	}
}

func (d *DockerUtild) FindContainerById(ctx context.Context, containerID string) (*types.Container, error) {
	// See:
	// https://docs.docker.com/engine/api/v1.41/#operation/ContainerList
	containerNameFilter := filters.Arg("id", containerID)
	containerFilter := filters.NewArgs(containerNameFilter)
	containerListOptions := types.ContainerListOptions{
		All:     true,
		Filters: containerFilter,
	}
	containers, err := d.client.ContainerList(ctx, containerListOptions)
	if err != nil {
		return nil, err
	}
	if len(containers) > 1 {
		return nil, fmt.Errorf("search container id %s but found more than 1 %w", containerID, SameNameContainerError)
	}
	return &containers[0], nil
}

func (d *DockerUtild) FindContainerByName(ctx context.Context, containerName string) (*types.Container, error) {
	// See:
	// https://docs.docker.com/engine/api/v1.41/#operation/ContainerList
	containerNameFilter := filters.Arg("name", containerName)
	containerFilter := filters.NewArgs(containerNameFilter)
	containerListOptions := types.ContainerListOptions{
		All:     true,
		Filters: containerFilter,
	}
	containers, err := d.client.ContainerList(ctx, containerListOptions)
	if err != nil {
		return nil, err
	}
	if len(containers) > 1 {
		return nil, fmt.Errorf("search container name %s but found more than 1 %w", containerName, SameNameContainerError)
	}
	return &containers[0], nil
}

func (d *DockerUtild) StartContainer(ctx context.Context, containerID string) error {
	// See:
	// https://docs.docker.com/engine/api/v1.41/#operation/ContainerStart
	return d.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

func (d *DockerUtild) StopContainer(ctx context.Context, containerID string) error {
	// See:
	// https://docs.docker.com/engine/api/v1.41/#operation/ContainerStop
	timeout := time.Second * 10
	return d.client.ContainerStop(ctx, containerID, &timeout)
}

func (d *DockerUtild) RemoveContainer(ctx context.Context, containerID string) error {
	// See:
	// https://docs.docker.com/engine/api/v1.41/#operation/ContainerDelete
	return d.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
}

func (d *DockerUtild) StopAndRemoveContainer(ctx context.Context, containerID string) {
	timeout := time.Second * 10
	_ = d.client.ContainerStop(ctx, containerID, &timeout)
	err := d.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
	if err != nil {
		log.Printf("StopAndRemoveContainer %s error: %s\n", d.GetShortId(containerID), err.Error())
	} else {
		log.Printf("StopAndRemoveContainer %s succeed\n", d.GetShortId(containerID))
	}
}

func (d *DockerUtild) GetContainerLog(ctx context.Context, containerID string, options types.ContainerLogsOptions) []string {
	// See :
	// https://docs.docker.com/engine/api/v1.41/#operation/ContainerLogs
	// and docker_api_logs_test.go
	reader, err := d.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		log.Printf("GetContainerLog %s error: %s\n", d.GetShortId(containerID), err.Error())
		return nil
	}
	defer reader.Close()

	actualStdout := new(bytes.Buffer)
	actualStderr := ioutil.Discard
	_, err = stdcopy.StdCopy(actualStdout, actualStderr, reader)
	if err != nil {
		log.Printf("GetContainerLog %s error: %s\n", d.GetShortId(containerID), err.Error())
		return nil
	}

	return strings.Split(actualStdout.String(), "\n")
}

func (d *DockerUtild) PrintLogByPoll(ctx context.Context, containerId string, secondInterval int) {
	since := "0"
	startIndex := len("2006-01-02T15:04:05.999999999Z")

	for {
		select {
		case <-ctx.Done():
			break
		default:
			var logs []string
			var err error

			logs, since, err = d.fetchLog(ctx, containerId, since)
			if err != nil {
				fmt.Println("Fetch log got error: " + err.Error())
				break
			} else {
				logs = logs[:len(logs)-2] //remove that last hard return line
				fmt.Printf("== Fetched log after %s\n", since)
				for _, log := range logs {
					fmt.Println(log[startIndex:])
				}
			}
		}

		time.Sleep(time.Duration(secondInterval) * time.Second)
	}
}

func (d *DockerUtild) fetchLog(ctx context.Context, containerId string, nextSince string) ([]string, string, error) {
	var option types.ContainerLogsOptions

	if nextSince == "0" {
		option = types.ContainerLogsOptions{
			ShowStdout: true,
			Timestamps: true,
		}
	} else {
		option = types.ContainerLogsOptions{
			ShowStdout: true,
			Timestamps: true,
			Since:      nextSince,
		}
	}

	currentRoundFetchedLogs := d.GetContainerLog(ctx, containerId, option)
	nextSince, err := getNextSince(currentRoundFetchedLogs)
	if err != nil {
		return make([]string, 0), "-1", err
	}

	return currentRoundFetchedLogs, nextSince, nil
}

func getNextSince(logs []string) (string, error) {
	if len(logs) == 0 {
		return "0", nil
	}

	lastLog := logs[len(logs)-2] // pay attention, last line is logs[len(logs)-1], it is always hard return line
	t, err := time.Parse(time.RFC3339Nano, strings.Split(lastLog, " ")[0])
	if err != nil {
		return "", err
	}
	nextSince := t.Format(time.RFC3339Nano)
	return nextSince, nil
}
