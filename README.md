
# Deploy old

Create a bucket to store cloud build -> jarviews-cloudbuild

Create a deploy token in gitlab for repository :
- Settings > Repository > Deploy Tokens

Put the token in a file -> deploykey.txt. Caution to don't add an endline char at the end of the file

Cypher the deploy token with a gcloud kms key and deploy the function :

```
# Create a keyrings:
gcloud kms keyrings create secrets --location=europe-west1
# Add a key in the keyring:
gcloud kms keys create gitlab --location=europe-west1 --keyring=secrets --purpose=encryption
# Encrypt the deploykey:
gcloud kms encrypt --location=europe-west1 --keyring=secrets --key gitlab --plaintext-file=deploykey.txt --ciphertext-file=deploykey.enc
# Authorize appspot (function user to use the key)
gcloud kms keys add-iam-policy-binding gitlab --location=europe-west1 --keyring=secrets --member=serviceAccount:wired-balm-258119@appspot.gserviceaccount.com --role roles/cloudkms.cryptoKeyDecrypter
gcloud functions deploy CreateBuildFromGitLab --runtime go111 --trigger-http --region=europe-west1 --env-vars-file env.yaml
```

Create a webhooks in gitlab with function url

gcloud projects add-iam-policy-binding wired-balm-258119 --member serviceAccount:840720119725@cloudservices.gserviceaccount.com --role roles/storage.admin
gcloud projects remove-iam-policy-binding wired-balm-258119 --member serviceAccount:840720119725@cloudservices.gserviceaccount.com --role roles/storage.admin


# Deploy

gcloud services enable runtimeconfig.googleapis.com

gcloud deployment-manager deployments create gitlab-hook --config deploy.yaml

-- TODO : add authorization by deployment

gcloud kms keys add-iam-policy-binding gitlab-hook \
  --location=europe-west1 --keyring=google-builds-gitlab \
  --member=serviceAccount:gitlab-hook@wired-balm-258119.iam.gserviceaccount.com \
  --role roles/cloudkms.cryptoKeyEncrypterDecrypter
  
gcloud projects add-iam-policy-binding wired-balm-258119  \
  --member=serviceAccount:gitlab-hook@wired-balm-258119.iam.gserviceaccount.com \
  --role roles/cloudbuild.builds.builder

gcloud projects add-iam-policy-binding wired-balm-258119  \
  --member=serviceAccount:gitlab-hook@wired-balm-258119.iam.gserviceaccount.com \
  --role roles/runtimeconfig.admin

gcloud functions deploy GitHookConfigHandler \
  --service-account=gitlab-hook@wired-balm-258119.iam.gserviceaccount.com \
  --runtime go111 --trigger-http --region=europe-west1 --env-vars-file env.yaml
  
gcloud functions deploy GitHookHandler \
  --service-account=gitlab-hook@wired-balm-258119.iam.gserviceaccount.com \
  --runtime go111 --trigger-http --region=europe-west1 --env-vars-file env.yaml

# Undeploy


- Bucket can only be deleted when they are empty
- Cloud KMS KeyRings and CryptoKeys cannot be deleted, so you need to abandon the resources.

```
gsutil rm -r gs://gitlab-hook-bucket/
gcloud deployment-manager deployments delete gitlab-hook
gcloud deployment-manager deployments delete gitlab-hook --delete-policy=abandon
``` 

# Environement variables to run locally 

gcloud iam service-accounts keys create ~/Documents/private/gitlab-hook-sa.json --iam-account gitlab-hook@wired-balm-258119.iam.gserviceaccount.com

go build && \
  PORT=8081 \
  GOOGLE_APPLICATION_CREDENTIALS=~/Documents/private/gitlab-hook-sa.json \
  REGION=europe-west1 \
  DEPLOYMENT_NAME=gitlab-hook \
  GCP_PROJECT=$(gcloud config get-value project) \
  ./cmd

curl -X POST localhost:8081/config -H "Content-Type:application/json" -d '{"project_id":42001,"username":"monUseName2", "deploy_token":"fsdfdsfdsezrzea444"}'


# TODO:

- Better README ^
- Username of deploy_token is hardcoded
- template of env.yaml
- Use service account instead of appspot account
- Automatically remove endl from gitlab token
- Better error management
- Bucket name as environement variables
- Tgz filen^ame with a timestamp or a uniq id (commit id ?)
- Don't tar .git folder
- use panic for error : cf https://github.com/GoogleCloudPlatform/golang-samples/blob/master/functions/tips/error.go ??
- runtimeconfig.admin role only on the gitlabhook config object not at the project level




gcloud kms keys add-iam-policy-binding gitlab --location=europe-west1 --keyring=secrets --member=serviceAccount:840720119725-compute@developer.gserviceaccount.com --role roles/cloudkms.cryptoKeyEncrypterDecrypter
