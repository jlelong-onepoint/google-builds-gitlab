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
	ProjectId uint `json:"project_id"`
	Project project `json:"project"`
}

type project struct {
	HttpUrl string `json:"http_url"`
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

	deployToken, err := pkg.Decrypt(GetKeyName(), repositoryConfig.EncryptedDeployToken)
	if err != nil {
		panic(err)
	}

	fmt.Println("Checkout sources")
	pkg.Checkout(pushHook.Project.HttpUrl, checkoutFolder, repositoryConfig.Username, *deployToken)

	fmt.Println("Tgz sources to bucket")
	bucketName := os.Getenv("DEPLOYMENT_NAME")
	err = sourceToBucket(bucketName)
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
			Object: "source.tgz",
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

	fmt.Println("Done")

}


func sourceToBucket(bucketName string) (err error){
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	writer := client.Bucket(bucketName).Object("source.tgz").NewWriter(ctx)
	defer func() {
		err = writer.Close()
	}()

	pkg.TarFolder(checkoutFolder, writer)

	return nil
}

/*
func getGitAccessKey() []byte {

	ctx := context.Background()
	client, err := cloudkms.NewKeyManagementClient(ctx)
	CheckIfError("create kms service client", err)

	accessKey, err := base64.StdEncoding.DecodeString(os.Getenv("GITLAB_ACCESS_TOKEN"))
	CheckIfError("decode access token from env var", err)

	// Build the request.
	req := &kmspb.DecryptRequest{
		Name:       "projects/wired-balm-258119/locations/europe-west1/keyRings/secrets/cryptoKeys/gitlab",
		Ciphertext: accessKey,
	}

	// Call the API.
	resp, err := client.Decrypt(ctx, req)
	CheckIfError("decrypt access token", err)

	return resp.Plaintext


}


 */

