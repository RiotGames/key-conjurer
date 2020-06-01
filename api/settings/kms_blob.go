package settings

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
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

	config := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(settings.AwsRegion)}))

	kmsClient := kms.New(config)

	blob, err := base64.StdEncoding.DecodeString(encryptedSettings)

	input := &kms.DecryptInput{
		CiphertextBlob: blob,
	}

	result, err := kmsClient.Decrypt(input)
	if err != nil {
		logger.Error("aws client failed to decrypt reason: ", err.Error())
		logger.Fatal(err.Error())
		panic(err)
	}

	if err := json.Unmarshal(result.Plaintext, &settings); err != nil {
		logger.Error("unable to unmarshal reason: ", err.Error())
		logger.Fatal(err.Error())
		panic(err)
	}

	return settings
}
