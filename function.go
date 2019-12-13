package google_builds_gitlab

import (
	cloudkms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"google.golang.org/api/cloudbuild/v1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"gopkg.in/src-d/go-git.v4"
	gitHttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"net/http"
	"os"
)

const checkoutFolder = "/tmp/project"
const cloudBuildYaml = checkoutFolder + "/cloudbuild.yaml"

type gitlabPushHook struct {
	ObjectKind string `json:"object_kind"`
	Project project `json:"project"`
}

type project struct {
    HttpUrl string `json:"http_url"`
}

func CreateBuildFromGitLab(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving request")

	var pushHook gitlabPushHook
	err := json.NewDecoder(r.Body).Decode(&pushHook)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}


	checkout(pushHook.Project.HttpUrl)

	ctx := context.Background()
	tar2GoogleStorage(ctx)

	cloudbuildYaml, err := ioutil.ReadFile(cloudBuildYaml)
	CheckIfError("read steps definition file", err)

	var build cloudbuild.Build
	err = yaml.Unmarshal(cloudbuildYaml, &build)
	CheckIfError("unmarshal steps definition file", err)

	build.Source = &cloudbuild.Source{
		StorageSource: &cloudbuild.StorageSource{
			Bucket: "jarviews-cloudbuild",
			Object: "source.tgz",
		},
	}

	cloudbuildService, err := cloudbuild.NewService(ctx)
	CheckIfError("create cloudbuild service", err)

	projectBuildService := cloudbuild.NewProjectsBuildsService(cloudbuildService)
	createCall := projectBuildService.Create("wired-balm-258119", &build)
	_, err = createCall.Do()
	CheckIfError("create new build", err)

}

func tar2GoogleStorage(ctx context.Context) {
	client, err := storage.NewClient(ctx)
	CheckIfError("create storage service client", err)

	writer := client.Bucket("jarviews-cloudbuild").Object("source.tgz").NewWriter(ctx)
	defer writer.Close()

	tarFolder(checkoutFolder, writer)
}

func checkout(gitUrl string) {
	fmt.Println("checkout: " + gitUrl)

	deployKey := string(getGitAccessKey())
	fmt.Println("Try to checkout with access key : [" + deployKey + "]")

	git, err := git.PlainClone(checkoutFolder, false, &git.CloneOptions{
		Auth: &gitHttp.BasicAuth{
			Username: "gitlab+deploy-token-67",
			Password: deployKey,
		},
		URL: gitUrl,
	})
	CheckIfError("clone repository", err)

	// ... retrieving the branch being pointed by HEAD
	ref, err := git.Head()
	CheckIfError("extract head", err)
	// ... retrieving the commit object
	commit, err := git.CommitObject(ref.Hash())
	CheckIfError("extract commit", err)

	fmt.Print(commit)
}

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

func CheckIfError(step string, err error) {
	if err == nil {
		fmt.Println(step)
		return
	}

	fmt.Printf("Error : %s %s\n", step, err)
	os.Exit(1) // TODO : Better error management
}




