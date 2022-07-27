package dockerUtild

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/myplayground-golang/fileUtild"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
)

/*

	image related operation functions

*/

func (d *DockerUtild) FindImageById(ctx context.Context, imageId string) (types.ImageSummary, error) {
	imageSummaries, err := d.client.ImageList(ctx, types.ImageListOptions{
		All: false,
	})
	if err != nil {
		panic(err)
	}

	for _, summary := range imageSummaries {
		if d.GetShortId(summary.ID) == imageId {
			return summary, nil
		}
	}

	return types.ImageSummary{}, errors.New("unable to find docker image by id" + imageId)
}

func (d *DockerUtild) FindImage(ctx context.Context, imageName string, imageTag string) (*[]types.ImageSummary, error) {
	searchResult := make([]types.ImageSummary, 0)

	// See:
	// https://docs.docker.com/engine/api/v1.41/#operation/ImageList
	// for what filters it should be set
	imageNameTagFilter := filters.Arg("reference", fmt.Sprintf("%s:%s", imageName, imageTag))
	imageFilter := filters.NewArgs(imageNameTagFilter)
	imageListOptions := types.ImageListOptions{
		All:     true,
		Filters: imageFilter,
	}
	images, err := d.client.ImageList(ctx, imageListOptions)
	if err != nil {
		return &searchResult, nil
	}

	for _, image := range images {
		searchResult = append(searchResult, image)
	}

	if len(searchResult) == 0 {
		return &searchResult, fmt.Errorf("image %s:%s. %w", imageName, imageTag, NoImageFoundError)
	} else if len(searchResult) > 1 {
		return &searchResult, fmt.Errorf("image %s:%s. %w", imageName, imageTag, SearchMultipleImageError)
	} else {
		return &searchResult, nil
	}
}

func (d *DockerUtild) BuildImage(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	// See:
	// https://docs.docker.com/engine/api/v1.41/#operation/ImageBuild
	// https://blog.csdn.net/Azj12345/article/details/121778095

	return d.client.ImageBuild(ctx, buildContext, options)
}

func (d *DockerUtild) BuildImageByDefaultOption(ctx context.Context, buildContext io.Reader, imageName string, imageTag string) (types.ImageBuildResponse, error) {
	tags := make([]string, 0)
	tag := fmt.Sprintf("%s:%s", imageName, imageTag)
	tags = append(tags, tag)

	options := types.ImageBuildOptions{
		Dockerfile:     "Dockerfile", // it is relative path in the input tar, the input tar is buildContext
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
		Tags:           tags,
	}

	return d.BuildImage(ctx, buildContext, options)
}

func (d *DockerUtild) BuildImageWithContextFolderByDefaultOption(ctx context.Context, contextFolder string, imageName string, imageTag string) (types.ImageBuildResponse, error) {
	currentWorkingFolder := fileUtild.GetCurrentWorkingFolder()
	tarFilePath, err := fileUtild.TarWithBase(contextFolder, currentWorkingFolder, ".")
	if err != nil {
		return types.ImageBuildResponse{}, err
	}

	dockerBuildContext, _ := os.Open(tarFilePath)

	defer func() {
		dockerBuildContext.Close()
		os.Remove(tarFilePath)
	}()

	return d.BuildImageByDefaultOption(ctx, dockerBuildContext, imageName, imageTag)
}

func (d *DockerUtild) BuildImageWithContextFolderByDefaultOptionAndOutput(ctx context.Context, contextFolder string, imageName string, imageTag string) (string, error) {
	buildResponse, err := d.BuildImageWithContextFolderByDefaultOption(ctx, contextFolder, imageName, imageTag)
	if err != nil {
		return "", err
	}
	response, err := ioutil.ReadAll(buildResponse.Body)
	if err != nil {
		return "", err
	}

	return string(response), nil
}
