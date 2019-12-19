# Deploy

Set environment variables :
```shell script
export GCP_PROJECT=$(gcloud config get-value project)
export GCP_PROJECT_NUMBER=$(gcloud projects describe $GCP_PROJECT --format="value(projectNumber)")
```

Enable needed services in the projet :
```shell script
gcloud services enable runtimeconfig.googleapis.com
gcloud services enable deploymentmanager.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable cloudkms.googleapis.com
```


Create the deployment :

```shell script
gcloud projects add-iam-policy-binding $GCP_PROJECT \
  --member serviceAccount:$GCP_PROJECT_NUMBER@cloudservices.gserviceaccount.com --role roles/storage.admin
gcloud deployment-manager deployments create gitlab-hook --template deploy.py --properties region:europe-west1
```

Add needed supplementary roles to created service account :
```shell script
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
```

Deploy githook functions (No easy way to deploy a function using google deployment)
```shell script
gcloud functions deploy GitHookConfigHandler \
  --service-account=gitlab-hook@wired-balm-258119.iam.gserviceaccount.com \
  --runtime go111 --trigger-http --region=europe-west1 --env-vars-file env.yaml
  
gcloud functions deploy GitHookHandler \
  --service-account=gitlab-hook@wired-balm-258119.iam.gserviceaccount.com \
  --runtime go111 --trigger-http --region=europe-west1 --env-vars-file env.yaml
```

# Configure a gitlab repository

Create a deploy token in gitlab for repository :
- Settings > Repository > Deploy Tokens

Create a webhooks in gitlab for push notification filling GitHookHandler function url


# Undeploy

- Bucket can only be deleted when they are empty
- Cloud KMS KeyRings and CryptoKeys cannot be deleted, so you need to abandon the resources.

```
gsutil rm -r gs://gitlab-hook-bucket/
gcloud deployment-manager deployments delete gitlab-hook
gcloud deployment-manager deployments delete gitlab-hook --delete-policy=abandon
gcloud projects remove-iam-policy-binding wired-balm-258119 --member serviceAccount:840720119725@cloudservices.gserviceaccount.com --role roles/storage.admin
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

- Clean README ^
- Checkout gitref based on hook informations
- Add deploy output for serviceAccount name and needed informations for subsequent commands
- Better error management
  - use panic for error ?? : cf https://github.com/GoogleCloudPlatform/golang-samples/blob/master/functions/tips/error.go ??
- Tgz filename with a timestamp or a uniq id (commit id ?)
- Don't tar .git folder
- runtimeconfig.admin role only on the gitlabhook config object not at the project level
- add authorization by deployment, not via command line
- Move the region to runtimeconfig at deployment time and remove the env.yaml file
