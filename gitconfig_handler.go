package google_builds_gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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

	log.Print("Adding a gitlab configuration...")

	// TODO : Switch to runtimeConfig object for configuration storage
	// TODO : Check http verb
	// TODO : Allow to delete a configuration
	// TODO : Allow to update a configuration

	configRequest, err := extractConfigRequest(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	encryptedDeployToken, err := pkg.Encrypt(GetKeyName(), configRequest.DeployToken)
	if err != nil {
		panic(err)
	}

	config := pkg.RepositoryConfig{
		GitProjectId:         *configRequest.ProjectId,
		Username:             configRequest.Username,
		EncryptedDeployToken: encryptedDeployToken,
	}

	err = pkg.NewConfigurationService(os.Getenv("GCP_PROJECT"), os.Getenv("DEPLOYMENT_NAME")).StoreConfig(config)
	if err != nil {
		panic(err)
	}

}

// "projects/wired-balm-258119/locations/europe-west1/keyRings/google-builds-gitlab/cryptoKeys/gitlab"
func GetKeyName() string {
	return fmt.Sprintf("projects/%v/locations/%v/keyRings/%v/cryptoKeys/%v",
		os.Getenv("GCP_PROJECT"), os.Getenv("REGION"), keyRingName, os.Getenv("DEPLOYMENT_NAME"))

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


