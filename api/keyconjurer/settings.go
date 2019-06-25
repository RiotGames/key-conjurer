package keyconjurer

import (
	"log"
	"os"
)

// Settings is used to hold keyconjurer settings
type Settings struct {
	AwsRegion              string
	AwsKMSKeyID            string `json:"awsKmsKeyId"`
	OneLoginReadUserID     string `json:"oneLoginReadUserId"`
	OneLoginReadUserSecret string `json:"oneLoginReadUserSecret"`
	OneLoginSamlID         string `json:"oneLoginSamlId"`
	OneLoginSamlSecret     string `json:"oneLoginSamlSecret"`
	OneLoginShard          string `json:"oneLoginShard"`
	OneLoginSubdomain      string `json:"oneLoginSubdomain"`
}

// NewSettings decrypts the encrypted settings then returns a new
//  Settings struct.
func NewSettings(aws *AWSClient, awsRegion string) *Settings {
	encryptedSettings := os.Getenv("EncryptedSettings")
	settings := &Settings{AwsRegion: awsRegion}
	err := aws.Decrypt(encryptedSettings, settings)
	if err != nil {
		log.Fatal("Could not load settings")
	}
	return settings
}
