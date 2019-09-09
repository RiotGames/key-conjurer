package settings

import (
	"log"
	"os"

	"keyconjurer-lambda/keyconjurer/awsclient"

	"github.com/sirupsen/logrus"
)

func init() {
	registerRetriever("kms_blob", NewSettingsFromKMSBlob)
}

// NewSettingsFromKMSBlob decrypts the encrypted settings in environment variable
// EncryptedSettings then returns a new Settings struct.
func NewSettingsFromKMSBlob(logger *logrus.Entry) *Settings {
	awsRegion := os.Getenv("AWSRegion")
	encryptedSettings := os.Getenv("EncryptedSettings")

	settings := &Settings{AwsRegion: awsRegion}

	aws := awsclient.NewAWSClient(awsRegion, logger)
	if err := aws.Decrypt(encryptedSettings, settings); err != nil {
		log.Fatal("Could not load settings")
	}
	return settings
}
