package dockerUtild

import (
	"bufio"
	"fmt"
	"github.com/myplayground-golang/fileUtild"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Initialize() error {
	return InitializeWithCurrentFolder("")
}

func InitializeWithCurrentFolder(currentFolder string) error {
	var err error

	// viper initial
	DockerConfigYamlReader = viper.New()
	//fmt.Println(GetCurrentWorkingFolder())

	if currentFolder == "" {
		currentFolder = GetCurrentWorkingFolder()
	}

	env := filepath.Join(currentFolder, ".env")

	if !fileUtild.IsFileExist(env) {
		DockerConfigYamlReader.AddConfigPath(currentFolder)
		DockerConfigYamlReader.SetConfigName("docker-compose")
		DockerConfigYamlReader.SetConfigType("yaml")
	} else { // exist .env, needs to do replacement in docker-compose.yaml
		source := filepath.Join(currentFolder, "docker-compose.yaml")
		target := filepath.Join(currentFolder, "docker-compose-env.yaml")
		fileUtild.CopyFile(source, target)

		// read the copied docker-compose-env.yaml file
		targetYamlContent, _ := ioutil.ReadFile(target)

		// read .env
		envFile, err := os.Open(env)
		if err != nil {
			panic(err)
		}
		defer envFile.Close()

		scanner := bufio.NewScanner(envFile)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.Split(line, "=")
			key := parts[0]
			value := parts[1]

			// do replace
			keyRegExp := regexp.MustCompile(fmt.Sprintf("\\${%s}", key))
			targetYamlContent = keyRegExp.ReplaceAll(targetYamlContent, []byte(value))
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}

		ioutil.WriteFile(target, targetYamlContent, 0666)

		DockerConfigYamlReader.AddConfigPath(currentFolder)
		DockerConfigYamlReader.SetConfigName("docker-compose-env")
		DockerConfigYamlReader.SetConfigType("yaml")
	}

	if err = DockerConfigYamlReader.ReadInConfig(); err != nil {
		panic(err)
	} else {
		fmt.Println("Viper read docker-compose.yaml successfully")
	}

	YamlServicesMap = getServiceConfigFromYaml(DockerConfigYamlReader)
	//fmt.Println(YamlServicesMap)

	return nil
}
