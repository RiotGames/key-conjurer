package settings

import (
	"log"
	"os"

	vault "github.com/riotgames/vault-go-client"

	"github.com/sirupsen/logrus"
)

func init() {
	registerRetriever("vault", NewSettingsFromVault)
}

func getVaultConfig() map[string]string {
	vaultSettings := map[string]string{
		"vaultRoleName":        os.Getenv("VAULT_ROLE_NAME"),
		"vaultSecretMountPath": os.Getenv("VAULT_SECRET_MOUNT_PATH"),
		"vaultSecretPath":      os.Getenv("VAULT_SECRET_PATH"),
		"vaultAWSAuthPath":     os.Getenv("VAULT_AWS_AUTH_PATH")}

	for key, value := range vaultSettings {
		if value == "" {
			log.Fatalf("%s was not set", key)
		}
	}
	return vaultSettings
}

// NewSettingsFromVault pulls configuration from a Vault instance
//  located at VAULT_ADDR
func NewSettingsFromVault(logger *logrus.Entry) *Settings {
	awsRegion := os.Getenv("AWSRegion")
	vaultConfig := getVaultConfig()

	settings := &Settings{AwsRegion: awsRegion}
	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		log.Fatal("Unable to get Vault Client")
	}

	if _, err := client.Auth.IAM.Login(vault.IAMLoginOptions{
		Role:      vaultConfig["vaultRoleName"],
		MountPath: vaultConfig["vaultAWSAuthPath"]}); err != nil {
		log.Fatal("Unable to login to Vault")
	}

	if _, err := client.KV2.Get(vault.KV2GetOptions{
		MountPath:     vaultConfig["vaultSecretMountPath"],
		SecretPath:    vaultConfig["vaultSecretPath"],
		UnmarshalInto: settings}); err != nil {
		log.Fatal("Unable to get vault settings from ", vaultConfig["vaultSecretMounthPath"]+"/"+vaultConfig["vaultSecretPath"])
	}
	return settings
}
