"""Creates a KMS key."""

def GenerateConfig(context):

  runtimeConfig = {
    "name": "runtimeConfig",
    "type": "runtimeconfig.v1beta1.config",
    "properties": {
      "config": context.env['deployment']
    }
  }

  service_account = {
    'name': 'serviceAccount',
    'type': 'iam.v1.serviceAccount',
    'properties': {
      'accountId': context.env['deployment'],
      'displayName': context.env['deployment']
    }
  }

  key_ring = {
    'name': 'keyRing',
    'type': 'gcp-types/cloudkms-v1:projects.locations.keyRings',
    'properties': {
      'parent': 'projects/' + context.env['project'] + '/locations/' + context.properties['region'],
      'keyRingId': 'google-builds-gitlab'
    }
  }

  crypto_key = {
    'name': 'cryptoKey',
    'type': 'gcp-types/cloudkms-v1:projects.locations.keyRings.cryptoKeys',
    'properties': {
      'parent': '$(ref.keyRing.name)',
      'cryptoKeyId': context.env['deployment'],
      'purpose': 'ENCRYPT_DECRYPT'
    },
  }

  bucket = {
    'name': context.env['project'] + '_' + context.env['deployment'], # The name of the deployment is the name of the bucket. Unless, accessControl failed
    'type': 'storage.v1.bucket',
    'properties': {
      'name': context.env['project'] + '_' + context.env['deployment'],
      'location': context.properties['region'],
    },
    'accessControl': {
      'gcpIamPolicy': {
        'bindings': [
          {
            'role': 'roles/storage.admin', # TODO : Reduce role, not sure this is always needed
            'members': [ 'serviceAccount:$(ref.serviceAccount.email)' ]
          }
        ]
      }
    }
  }

  resources = [runtimeConfig, service_account, key_ring, crypto_key, bucket]

  return { 'resources': resources }

