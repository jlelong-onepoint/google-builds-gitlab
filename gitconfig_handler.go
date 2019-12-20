package google_builds_gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"peekmoon.org/google-builds-gitlab/pkg"
)

const (
	keyRingName = "google-builds-gitlab"
)

type addRepositoryConfigRequest struct {
	ProjectId   *uint  `json:"project_id"`
	Username    string `json:"username"`
	DeployToken string `json:"deploy_token"`
}


func GitHookConfigHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Adding a gitlab configuration...")

	// TODO : Check http verb
	// TODO : Allow to delete a configuration
	// TODO : Allow to update a configuration

	configRequest, err := extractConfigRequest(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cypherService := pkg.NewCypherService(os.Getenv("GCP_PROJECT"), os.Getenv("REGION"), keyRingName, os.Getenv("DEPLOYMENT_NAME"))

	encryptedDeployToken := cypherService.Encrypt(configRequest.DeployToken)

	config := pkg.RepositoryConfig{
		GitProjectId:         *configRequest.ProjectId,
		Username:             configRequest.Username,
		EncryptedDeployToken: encryptedDeployToken,
	}

	pkg.NewConfigurationService(os.Getenv("GCP_PROJECT"), os.Getenv("DEPLOYMENT_NAME")).StoreConfig(config)
}

func extractConfigRequest(reader io.Reader) (addRepositoryConfigRequest, error) {

	var config addRepositoryConfigRequest

	if err := json.NewDecoder(reader).Decode(&config); err != nil {
		return config, err
	}

	if config.ProjectId == nil {
		return config, errors.New("Project id is not provided")
	}

	if config.Username == "" {
		return config, errors.New("Username is empty")
	}

	if config.DeployToken == "" {
		return config, errors.New("Deploy token is empty")
	}

	return config, nil
}


