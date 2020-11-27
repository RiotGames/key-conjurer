package settings

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

func init() {
	registerRetriever("kms_blob", NewSettingsFromKMSBlob)
}

// NewSettingsFromKMSBlob decrypts the encrypted settings in environment
//  variable EncryptedSettings then returns a new Settings struct.
func NewSettingsFromKMSBlob() (*Settings, error) {
	awsRegion := os.Getenv("AWSRegion")
	encryptedSettings := os.Getenv("EncryptedSettings")

	settings := &Settings{AwsRegion: awsRegion}

	config, err := session.NewSession(&aws.Config{Region: aws.String(settings.AwsRegion)})
	if err != nil {
		return nil, err
	}

	kmsClient := kms.New(config)

	blob, err := base64.StdEncoding.DecodeString(encryptedSettings)

	input := &kms.DecryptInput{
		CiphertextBlob: blob,
	}

	result, err := kmsClient.Decrypt(input)
	if err != nil {
		return nil, fmt.Errorf("aws client failed to decrypt: %w", err)
	}

	if err := json.Unmarshal(result.Plaintext, &settings); err != nil {
		return nil, fmt.Errorf("unable to marshal to JSON: %w", err)
	}

	return settings, nil
}
