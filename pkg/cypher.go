package pkg

import (
	cloudkms "cloud.google.com/go/kms/apiv1"
	"context"
	"fmt"
	"github.com/pkg/errors"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type CypherService struct {
	keyName string
}

func NewCypherService(projectId string, region string, keyRingName string, keyName string) CypherService {
	keyNameTemplate := "projects/%v/locations/%v/keyRings/%v/cryptoKeys/%v"
	fullKeyName := fmt.Sprintf(keyNameTemplate, projectId, region, keyRingName, keyName)
	return CypherService{ keyName : fullKeyName }
}


func (service CypherService) Encrypt(value string) []byte {
	ctx := context.Background()

	client, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Build the encrypt request.
	req := kmspb.EncryptRequest{
		Name:      service.keyName,
		Plaintext: []byte(value),
	}
	// Call the API.
	resp, err := client.Encrypt(ctx, &req)
	if err != nil {
		panic(errors.Wrapf(err, "Unable to encrypt token with key %v", service.keyName))
	}

	return resp.Ciphertext
}

func (service CypherService) Decrypt(value []byte) string {
	ctx := context.Background()
	client, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close()


	// Build the decrypt request.
	req := kmspb.DecryptRequest{
		Name: service.keyName,
		Ciphertext: value,
	}

	// Call the API.
	resp, err := client.Decrypt(ctx, &req)
	if err != nil {
		panic(errors.Wrapf(err, "Unable to decrypt token with key %v", service.keyName))
	}

	result := string(resp.Plaintext)

	return result
}
