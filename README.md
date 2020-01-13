# Deploy

Commands deploy in the current default projet. Billing must be activated on project.
Set default gcp project : 

```shell script
gcloud config set project <PROJECT_NAME>
```

Set environment variables :
```shell script
export GCP_PROJECT=$(gcloud config get-value project)
export GCP_PROJECT_NUMBER=$(gcloud projects describe $GCP_PROJECT --format="value(projectNumber)")
export DEPLOYMENT_NAME=gitlab-hook
export REGION=europe-west1
export SA_EMAIL=$DEPLOYMENT_NAME@$GCP_PROJECT.iam.gserviceaccount.com
```

Enable needed services in the project :
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
  --runtime go111 --trigger-http --region=$REGION --set-env-vars=DEPLOYMENT_NAME=$DEPLOYMENT_NAME,REGION=$REGION

gcloud functions deploy GitHookHandler \
  --service-account=$SA_EMAIL \
  --runtime go111 --trigger-http --region=$REGION --set-env-vars=DEPLOYMENT_NAME=$DEPLOYMENT_NAME,REGION=$REGION
```

# Configure a gitlab repository

Create a deploy token in gitlab for repository :
- Settings > Repository > Deploy Tokens

Create a webhooks in gitlab for push notification filling GitHookHandler function url.

Call configuration function with git token informations. git_project_id can be retrieve under menu Settings/general

```shell script
curl -X POST <function-config-url> -H "Content-Type:application/json" \
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
gcloud functions delete GitHookConfigHandler --region=$REGION
gcloud functions delete GitHookHandler --region=$REGION
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
curl -X POST localhost:8081 -H "Content-Type:application/json" -d '{"project_id":2333, "after":"99d54b9c1c29e20798cc4d45ceb1bfc47bd0d70b", "project" : { "git_http_url":"https://gitlab.groupeonepoint.com/cds-bdx/aventure/jarviews/repository-index-reader.git"}}'
```

# Projects expectation

Your project is expected to have a `cloudbuild.yaml` at the root of your project.


# TODO:

- Add deploy output for serviceAccount name and needed informations for subsequent commands
- runtimeconfig.admin role only on the gitlabhook config object not at the project level
- add authorization by deployment, not via command line
- Move the region to runtimeconfig at deployment time
- Add error when a hook is received on a projet without configuration
- Allow to have project from multiple gitlab (avoid project_id clash)
- Add a secret token from gitlab
- Move pkg to internal and move every file in it's own package ?
- Init client only in init method : https://cloud.google.com/functions/docs/concepts/go-runtime
- In Decrypt/Encrypt mutualize google service creation
