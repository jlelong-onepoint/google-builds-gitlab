# Deploy

Set environment variables :
```shell script
export GCP_PROJECT=$(gcloud config get-value project)
export GCP_PROJECT_NUMBER=$(gcloud projects describe $GCP_PROJECT --format="value(projectNumber)")
export DEPLOYMENT_NAME=gitlab-hook
export REGION=europe-west1
export SA_EMAIL=$DEPLOYMENT_NAME@$GCP_PROJECT.iam.gserviceaccount.com
```

Enable needed services in the projet :
```shell script
gcloud services enable runtimeconfig.googleapis.com
gcloud services enable deploymentmanager.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable cloudkms.googleapis.com
gcloud services enable cloudfunctions.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

Create the deployment :

```shell script
gcloud projects add-iam-policy-binding $GCP_PROJECT \
  --member serviceAccount:$GCP_PROJECT_NUMBER@cloudservices.gserviceaccount.com --role roles/storage.admin
gcloud deployment-manager deployments create $DEPLOYMENT_NAME --template deploy.py --properties region:$REGION
```

Add needed supplementary roles to created service account :
```shell script
gcloud kms keys add-iam-policy-binding $DEPLOYMENT_NAME \
  --location=$REGION --keyring=google-builds-gitlab \
  --member=serviceAccount:$SA_EMAIL \
  --role roles/cloudkms.cryptoKeyEncrypterDecrypter
  
gcloud projects add-iam-policy-binding $GCP_PROJECT  \
  --member=serviceAccount:$SA_EMAIL \
  --role roles/cloudbuild.builds.builder

gcloud projects add-iam-policy-binding $GCP_PROJECT  \
  --member=serviceAccount:$SA_EMAIL \
  --role roles/runtimeconfig.admin
```

Deploy githook functions (No easy way to deploy a function using google deployment)
```shell script
gcloud functions deploy GitHookConfigHandler \
  --service-account=$SA_EMAIL \
  --runtime go111 --trigger-http --region=$REGION --env-vars-file env.yaml
  
gcloud functions deploy GitHookHandler \
  --service-account=$SA_EMAIL \
  --runtime go111 --trigger-http --region=$REGION --env-vars-file env.yaml
```

# Configure a gitlab repository

Create a deploy token in gitlab for repository :
- Settings > Repository > Deploy Tokens

Create a webhooks in gitlab for push notification filling GitHookHandler function url

Call configuration function with git token informations. git_project_id can be retrieve under menu Settings/general

```shell script
curl -X POST <function-config-url>/config -H "Content-Type:application/json" \
  -d '{"project_id":<git_projet_id>,"username":"<deploy_token_username>", "deploy_token":"<deploy_token>"}'
```

# Undeploy

- Bucket can only be deleted when they are empty
- Cloud KMS KeyRings and CryptoKeys cannot be deleted, so you need to abandon the resources.

```shell script
gsutil rm -r gs://$GCP_PROJECT_$DEPLOYMENT_NAME/
gcloud deployment-manager deployments delete $DEPLOYMENT_NAME
gcloud deployment-manager deployments delete $DEPLOYMENT_NAME --delete-policy=abandon
gcloud projects remove-iam-policy-binding $GCP_PROJECT \
  --member serviceAccount:$GCP_PROJECT_NUMBER@cloudservices.gserviceaccount.com --role roles/storage.admin
``` 

# Environement variables to run/debug locally 

gcloud iam service-accounts keys create ~/Documents/private/gitlab-hook-sa.json --iam-account $DEPLOYMENT_NAME@$GCP_PROJECT.iam.gserviceaccount.com

```shell script
go build && \
  PORT=8081 \
  GOOGLE_APPLICATION_CREDENTIALS=~/Documents/private/gitlab-hook-sa.json \
  REGION=europe-west1 \
  DEPLOYMENT_NAME=$DEPLOYMENT_NAME \
  GCP_PROJECT=$(gcloud config get-value project) \
  ./cmd
```

Start a hook :
```shell script
curl -X POST localhost:8081 -H "Content-Type:application/json" -d '{"project_id":2333, "project" : { "http_url":"https://gitlab.groupeonepoint.com/cds-bdx/aventure/jarviews/repository-index-reader.git"}}'
```

# TODO:

- Checkout gitref based on hook informations
- Add deploy output for serviceAccount name and needed informations for subsequent commands
- Better error management
  - use panic for error ?? : cf https://github.com/GoogleCloudPlatform/golang-samples/blob/master/functions/tips/error.go ??
- Tgz filename with a timestamp or a uniq id (commit id ?)
- Don't tar .git folder
- runtimeconfig.admin role only on the gitlabhook config object not at the project level
- add authorization by deployment, not via command line
- Move the region to runtimeconfig at deployment time and remove the env.yaml file
