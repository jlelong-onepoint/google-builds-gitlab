package google_builds_gitlab

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
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

	repositoryConfig := pkg.NewConfigurationService(os.Getenv("GCP_PROJECT"), os.Getenv("DEPLOYMENT_NAME")).ReadConfig(pushHook.ProjectId)

	cypherService := pkg.NewCypherService(os.Getenv("GCP_PROJECT"), os.Getenv("REGION"), keyRingName, os.Getenv("DEPLOYMENT_NAME"))

	deployToken := cypherService.Decrypt(repositoryConfig.EncryptedDeployToken)

	fmt.Println("Checkout sources")
	pkg.Checkout(pushHook.Project.HttpUrl, pushHook.CommitSha1, checkoutFolder, repositoryConfig.Username, deployToken)

	defer func() {
		fmt.Println("Delete checkout folder") // Unless if session is reused, checkout failed
		if err := os.RemoveAll(checkoutFolder); err != nil {
			panic(errors.Wrapf(err, "Unable to delete folder %v", checkoutFolder))
		}
	}()

	fmt.Println("Tgz sources to bucket")
	bucketName := os.Getenv("GCP_PROJECT") + "_" + os.Getenv("DEPLOYMENT_NAME")
	tgzName := fmt.Sprintf("source-%d-%s.tgz", pushHook.ProjectId, pushHook.CommitSha1)

	sourceToBucket(bucketName, tgzName)

	fmt.Println("Read cloud build definition and prepare build object")
	cloudbuildYaml, err := ioutil.ReadFile(cloudBuildYaml)
	if err != nil {
		errorExit(w,"unable to read cloud build definition file: %v", err)
	}

	var build cloudbuild.Build
	err = yaml.Unmarshal(cloudbuildYaml, &build)
	if err != nil {
		errorExit(w,"unable to parse cloud build definition : %v", err)
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
		panic(err)
	}

	projectBuildService := cloudbuild.NewProjectsBuildsService(cloudbuildService)
	createCall := projectBuildService.Create(os.Getenv("GCP_PROJECT"), &build)
	_, err = createCall.Do()
	if err != nil {
		panic(errors.Wrapf(err,"Unable to create build from bucket %v and object %v", bucketName, tgzName))
	}

	fmt.Println("Done")

}


func errorExit(w http.ResponseWriter, template string, e error){
	msg := fmt.Sprintf(template, e)
	fmt.Println(msg)
	http.Error(w, msg, http.StatusBadRequest)
	panic(e)
}

func sourceToBucket(bucketName string, objectName string) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	writer := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	defer func() {
		if ferr := writer.Close(); ferr != nil {
			panic(errors.Wrapf(ferr, "Unable to store source in bucket: %v", bucketName))
		}
	}()

	pkg.TarFolder(checkoutFolder, writer, ".git", ".gitignore")
}


