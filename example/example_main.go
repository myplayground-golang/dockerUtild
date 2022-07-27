package main

import (
	"context"
	"flag"
	"fmt"
	du "github.com/myplayground-golang/dockerUtild"
	"strings"
	"time"
)

func init() {
	du.Initialize()
}

func main() {
	var serviceName = flag.String("service", "", "serviceName in docker-compose.yml")
	flag.Parse()

	if *serviceName == "" {
		panic("You need to specify the service name in docker-compose.yaml!")
	}

	ctx := context.Background()
	option := du.ParseCreateContainerOptionFromYaml(du.YamlServicesMap, *serviceName)

	dockerUtil := du.NewDockerUtild()
	containerId, err := dockerUtil.CreateContainerWithOption(ctx, option)
	if err != nil {
		panic(err)
	}
	fmt.Println("created container id:" + dockerUtil.GetShortId(containerId))

	err = dockerUtil.StartContainer(ctx, containerId)
	if err != nil {
		panic(err)
	}
	fmt.Println("started container id:" + dockerUtil.GetShortId(containerId))
	time.Sleep(time.Duration(1) * time.Minute)
	fetchLogCtx, cancelFetchLog := context.WithCancel(ctx)
	go dockerUtil.PrintLogByPoll(fetchLogCtx, containerId, 30)

	sleepCount := 0

	for {
		container, err := dockerUtil.FindContainerByName(ctx, option.ContainerName)
		if err != nil {
			fmt.Println("Find container by name got err: " + err.Error())
			break
		}

		if strings.Contains(container.Status, "Exited") {
			fmt.Println("cypress container status is Exited.")
			break
		}

		//fmt.Printf("Container status: %s. Waiting for cypress container to complete running...\n", container.Status)

		time.Sleep(time.Duration(1) * time.Minute)
		sleepCount = sleepCount + 1

		if sleepCount > 10 {
			fmt.Println("Waited too long, force to end up waiting.")
			break
		}

	}

	// stop fetching container log in goroutine
	cancelFetchLog()

	// cleanup the docker container
	dockerUtil.StopAndRemoveContainer(ctx, dockerUtil.GetShortId(containerId))
}
