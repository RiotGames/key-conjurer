package settings

import (
	"fmt"
	"os"

	vault "github.com/riotgames/vault-go-client"
)

type vaultConfig struct {
	RoleName        string `mapstructure:"vault.roleName"`
	AWSAuthPath     string `mapstructure:"vault.secretMountPath"`
	SecretMountPath string `mapstructure:"vault.secretPath"`
	SecretPath      string `mapstructure:"vault.awsAuthPath"`
}

func retrieveFromVault() (*Settings, error) {
	awsRegion := os.Getenv("AWSRegion")
	cfg := vaultConfig{
		RoleName:        os.Getenv("VAULT_ROLE_NAME"),
		SecretMountPath: os.Getenv("VAULT_SECRET_MOUNT_PATH"),
		SecretPath:      os.Getenv("VAULT_SECRET_PATH"),
		AWSAuthPath:     os.Getenv("VAULT_AWS_AUTH_PATH"),
	}

	settings := &Settings{AwsRegion: awsRegion}
	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("unable to get Vault client")
	}

	opts := vault.IAMLoginOptions{
		Role:      cfg.RoleName,
		MountPath: cfg.AWSAuthPath,
	}

	if _, err := client.Auth.IAM.Login(opts); err != nil {
		return nil, fmt.Errorf("unable to login to Vault")
	}

	kvOpts := vault.KV2GetOptions{
		MountPath:     cfg.SecretMountPath,
		SecretPath:    cfg.SecretPath,
		UnmarshalInto: settings,
	}

	if _, err := client.KV2.Get(kvOpts); err != nil {
		return nil, fmt.Errorf("Unable to get vault settings from %s", cfg.SecretMountPath+"/"+cfg.SecretPath)
	}

	return settings, nil
}
