package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/api/runtimeconfig/v1beta1"
)

type ConfigurationService struct {
	projectId string
	configName string
}

type RepositoryConfig struct {
	GitProjectId         uint   `json:"project_id"`
	Username             string `json:"username"`
	EncryptedDeployToken []byte `json:"encrypted_token"`
}

func NewConfigurationService(projectId string, configName string) *ConfigurationService {
	return &ConfigurationService{
		configName : "projects/" + projectId + "/configs/" + configName,
	}
}


func (service ConfigurationService) ReadConfig(projectId uint) RepositoryConfig {
	fmt.Printf("Retriving configuration for project %v\n", projectId)

	ctx := context.Background()

	runtimeConfigService, err := runtimeconfig.NewService(ctx)
	if err != nil {
		panic(err)
	}

	variableName := fmt.Sprintf("%v/variables/%d", service.configName, projectId)
	getVariable := runtimeconfig.NewProjectsConfigsVariablesService(runtimeConfigService).Get(variableName)

	variable, err := getVariable.Do()
	if err != nil {
		panic(errors.Wrapf(err, "Unable to get Variable : %v", variableName))
	}

	var config RepositoryConfig
	err = json.Unmarshal([]byte(variable.Text), &config)
	if  err != nil {
		panic(errors.Wrapf(err, "Unable to unmarshal configuration : %v", variable.Text))
	}

	return config
}


func (service ConfigurationService) StoreConfig(config RepositoryConfig) {
	fmt.Printf("Storing configuration for project %v\n", config.GitProjectId)

	ctx := context.Background()

	runtimeConfigService, err := runtimeconfig.NewService(ctx)
	if err != nil {
		panic(err)
	}

	text, err := json.Marshal(config)
	if  err != nil {
		panic(err)
	}

	variableName := fmt.Sprintf("%v/variables/%d", service.configName, config.GitProjectId)
	createVariable := runtimeconfig.NewProjectsConfigsVariablesService(runtimeConfigService).Create(service.configName, &runtimeconfig.Variable{
		Name: variableName,
		Text: string(text),
	} )

	_, err = createVariable.Do()
	if err != nil {
		panic(errors.Wrapf(err, "Unable to store config : %s <== %s", variableName, string(text)))
	}

	fmt.Printf("Variable created : %v\n", variableName)
}





