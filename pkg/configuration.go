package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/runtimeconfig/v1beta1"
	"log"
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


func (service ConfigurationService) ReadConfig(projectId uint) (*RepositoryConfig, error) {
	log.Printf("Retriving configuration for project %v", projectId)

	ctx := context.Background()

	runtimeConfigService, err := runtimeconfig.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("runtimeconfig.NewService: %v", err)
	}

	variableName := fmt.Sprintf("%v/variables/%d", service.configName, projectId)
	getVariable := runtimeconfig.NewProjectsConfigsVariablesService(runtimeConfigService).Get(variableName)

	variable, err := getVariable.Do()
	if err != nil {
		return nil, err
	}

	var config RepositoryConfig
	err = json.Unmarshal([]byte(variable.Text), &config)
	if  err != nil {
		return nil,err
	}

	log.Printf("Config read : %W", config)

	return &config, nil
}


func (service ConfigurationService) StoreConfig(config RepositoryConfig) (err error) {
	log.Printf("Storing configuration for project %v", config.GitProjectId)

	ctx := context.Background()

	runtimeConfigService, err := runtimeconfig.NewService(ctx)
	if err != nil {
		return fmt.Errorf("runtimeconfig.NewService: %v", err)
	}

	text, err := json.Marshal(config)
	if  err != nil {
		return err
	}

	variableName := fmt.Sprintf("%v/variables/%d", service.configName, config.GitProjectId)
	createVariable := runtimeconfig.NewProjectsConfigsVariablesService(runtimeConfigService).Create(service.configName, &runtimeconfig.Variable{
		Name: variableName,
		Text: string(text),
	} )

	variable, err := createVariable.Do()
	if err != nil {
		return err
	}

	log.Printf("Variable create : %W", variable)

	return err
}





