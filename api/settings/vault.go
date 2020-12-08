package settings

import (
	"fmt"
	"os"

	vault "github.com/riotgames/vault-go-client"

	"github.com/sirupsen/logrus"
)

func init() {
	registerRetriever("vault", NewSettingsFromVault)
}

func getVaultConfig() (map[string]string, error) {
	vaultSettings := map[string]string{
		"vaultRoleName":        os.Getenv("VAULT_ROLE_NAME"),
		"vaultSecretMountPath": os.Getenv("VAULT_SECRET_MOUNT_PATH"),
		"vaultSecretPath":      os.Getenv("VAULT_SECRET_PATH"),
		"vaultAWSAuthPath":     os.Getenv("VAULT_AWS_AUTH_PATH")}

	for key, value := range vaultSettings {
		if value == "" {
			return vaultSettings, fmt.Errorf("%s was not set", key)
		}
	}

	return vaultSettings, nil
}

// NewSettingsFromVault pulls configuration from a Vault instance
//  located at VAULT_ADDR
func NewSettingsFromVault(logger *logrus.Entry) (*Settings, error) {
	awsRegion := os.Getenv("AWSRegion")
	vaultConfig, err := getVaultConfig()
	if err != nil {
		return nil, err
	}

	settings := &Settings{AwsRegion: awsRegion}
	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("unable to get Vault client")
	}

	opts := vault.IAMLoginOptions{
		Role:      vaultConfig["vaultRoleName"],
		MountPath: vaultConfig["vaultAWSAuthPath"],
	}

	if _, err := client.Auth.IAM.Login(opts); err != nil {
		return nil, fmt.Errorf("unable to login to Vault")
	}

	kvOpts := vault.KV2GetOptions{
		MountPath:     vaultConfig["vaultSecretMountPath"],
		SecretPath:    vaultConfig["vaultSecretPath"],
		UnmarshalInto: settings,
	}

	if _, err := client.KV2.Get(kvOpts); err != nil {
		return nil, fmt.Errorf("Unable to get vault settings from %s", vaultConfig["vaultSecretMounthPath"]+"/"+vaultConfig["vaultSecretPath"])
	}

	return settings, nil
}
