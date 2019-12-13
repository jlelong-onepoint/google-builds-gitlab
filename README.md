
# Deploy

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

# Environement variables to run locally 

- PORT=8081
- GOOGLE_APPLICATION_CREDENTIALS=<account-key.json>
- GITLAB_ACCESS_TOKEN=<access-token-cypher>

# TODO:

- Better README ^
- Username of deploy_token is hardcoded
- template of env.yaml
- Use service account instead of appspot account
- Automatically remove endl from gitlab token
- Better error management
- Bucket name as environement variables
- Tgz filen^ame with a timestamp or a uniq id (commit id ?)
