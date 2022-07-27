package dockerUtild

import "os"

func GetCurrentWorkingFolder() string {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return currentWorkingDirectory
}
