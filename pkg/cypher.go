package pkg

import (
	cloudkms "cloud.google.com/go/kms/apiv1"
	"context"
	"fmt"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

func Encrypt(keyName string, value string) ([]byte, error) {
	ctx := context.Background()

	client, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudkms.NewKeyManagementClient: %v", err)
	}
	defer client.Close()

	// Build the encrypt request.
	req := kmspb.EncryptRequest{
		Name:      keyName,
		Plaintext: []byte(value),
	}
	// Call the API.
	resp, err := client.Encrypt(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("Encrypt: %v", err)
	}

	return resp.Ciphertext, nil

}

func Decrypt(keyName string, value []byte) (*string, error) {
	ctx := context.Background()
	client, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()


	// Build the decrypt request.
	req := kmspb.DecryptRequest{
		Name: keyName,
		Ciphertext: value,
	}

	// Call the API.
	resp, err := client.Decrypt(ctx, &req)
	if err != nil {
		return nil, err
	}

	result := string(resp.Plaintext)

	return &result, nil
}
