package main

import (
	"context"
	"fmt"
	du "github.com/myplayground-golang/dockerUtild"
	"github.com/myplayground-golang/fileUtild"
	"path/filepath"
)

func panicErrorIfNotNil(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var err error

	imageName := "dchen_test_build_image"
	imageTag := "1_trunk_07062022"

	dockerBuildContextFolder := filepath.Join(fileUtild.GetCurrentWorkingFolder(), "docker")
	dockerUtil := du.NewDockerUtild()
	response, err := dockerUtil.BuildImageWithContextFolderByDefaultOptionAndOutput(context.Background(), dockerBuildContextFolder, imageName, imageTag)
	panicErrorIfNotNil(err)
	fmt.Println(response)
}
