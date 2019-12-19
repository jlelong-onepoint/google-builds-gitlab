package google_builds_gitlab

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"google.golang.org/api/cloudbuild/v1"
	"io/ioutil"
	"net/http"
	"os"
	"peekmoon.org/google-builds-gitlab/pkg"
)


const (
	checkoutFolder = "/tmp/sources"
	cloudBuildYaml = checkoutFolder + "/cloudbuild.yaml"
)

type gitlabPushHook struct {
	ObjectKind string `json:"object_kind"`
	CommitSha1 string `json:"after"`
	ProjectId uint `json:"project_id"`
	Project project `json:"project"`
}

type project struct {
	HttpUrl string `json:"git_http_url"`
}

func GitHookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving git hook")

	var pushHook gitlabPushHook
	err := json.NewDecoder(r.Body).Decode(&pushHook)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read request : %v", err), http.StatusBadRequest)
		return
	}

	repositoryConfig, err := pkg.NewConfigurationService(os.Getenv("GCP_PROJECT"), os.Getenv("DEPLOYMENT_NAME")).ReadConfig(pushHook.ProjectId)
	if err != nil {
		panic(err)
	}

	cypherService := pkg.NewCypherService(os.Getenv("GCP_PROJECT"), os.Getenv("REGION"), keyRingName, os.Getenv("DEPLOYMENT_NAME"))

	deployToken, err := cypherService.Decrypt(repositoryConfig.EncryptedDeployToken)
	if err != nil {
		panic(err)
	}

	fmt.Println("Checkout sources")
	err = pkg.Checkout(pushHook.Project.HttpUrl, pushHook.CommitSha1, checkoutFolder, repositoryConfig.Username, *deployToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to checkout sources: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("Tgz sources to bucket")
	bucketName := os.Getenv("GCP_PROJECT") + "_" + os.Getenv("DEPLOYMENT_NAME")
	tgzName := fmt.Sprintf("source-%d-%s.tgz", pushHook.ProjectId, pushHook.CommitSha1)

	err = sourceToBucket(bucketName, tgzName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to store source in bucket: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("Read cloud build definition and prepare build object")
	cloudbuildYaml, err := ioutil.ReadFile(cloudBuildYaml)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read cloud build definition file: %v", err), http.StatusBadRequest)
		return
	}

	var build cloudbuild.Build
	err = yaml.Unmarshal(cloudbuildYaml, &build)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to parse cloud build definition : %v", err), http.StatusInternalServerError)
		return
	}

	build.Source = &cloudbuild.Source{
		StorageSource: &cloudbuild.StorageSource{
			Bucket: bucketName,
			Object: tgzName,
		},
	}

	fmt.Println("Submits builds")
	cloudbuildService, err := cloudbuild.NewService(context.Background())
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to create cloud build service : %v", err), http.StatusInternalServerError)
		return
	}

	projectBuildService := cloudbuild.NewProjectsBuildsService(cloudbuildService)
	createCall := projectBuildService.Create(os.Getenv("GCP_PROJECT"), &build)
	_, err = createCall.Do()
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to submit builds : %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("Delete checkout folder") // Unless if session is reused, checkout failed
	os.RemoveAll(checkoutFolder)

	fmt.Println("Done")

}


func sourceToBucket(bucketName string, objectName string) (err error){
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	writer := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	defer func() {
		err = writer.Close()
	}()

	pkg.TarFolder(checkoutFolder, writer)

	return nil
}


