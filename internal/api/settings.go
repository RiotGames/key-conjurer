package api

import (
	"context"
	"fmt"
	"os"

	"github.com/riotgames/vault-go-client"
)

// Settings is used to hold keyconjurer settings
type Settings struct {
	OktaHost  string `json:"oktaHost"`
	OktaToken string `json:"oktaToken"`
}

var SettingsProviders = map[string]SettingsProvider{}

func init() {
	SettingsProviders["env"] = SettingsProviderFunc(RetrieveSettingsFromEnv)
	SettingsProviders["vault"] = VaultRetriever{
		RoleName:        os.Getenv("VAULT_ROLE_NAME"),
		SecretMountPath: os.Getenv("VAULT_SECRET_MOUNT_PATH"),
		SecretPath:      os.Getenv("VAULT_SECRET_PATH"),
		AWSAuthPath:     os.Getenv("VAULT_AWS_AUTH_PATH"),
	}
}

type SettingsProvider interface {
	FetchSettings(ctx context.Context) (*Settings, error)
}

type SettingsProviderFunc func(ctx context.Context) (*Settings, error)

func (fn SettingsProviderFunc) FetchSettings(ctx context.Context) (*Settings, error) {
	return fn(ctx)
}

func NewSettings(ctx context.Context) (*Settings, error) {
	prov := "vault"
	if nextProv, ok := os.LookupEnv("SETTINGS_PROVIDER"); ok {
		prov = nextProv
	}

	entry, ok := SettingsProviders[prov]
	if !ok {
		return nil, fmt.Errorf("no settings provider with the name %q", prov)
	}

	return entry.FetchSettings(ctx)
}

func RetrieveSettingsFromEnv(_ context.Context) (*Settings, error) {
	s := Settings{
		OktaHost:  os.Getenv("OKTA_HOST"),
		OktaToken: os.Getenv("OKTA_TOKEN"),
	}

	return &s, nil
}

type VaultRetriever struct {
	RoleName        string
	AWSAuthPath     string
	SecretMountPath string
	SecretPath      string
}

func (v VaultRetriever) FetchSettings(_ context.Context) (*Settings, error) {
	var settings Settings
	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("unable to get Vault client: %w", err)
	}

	opts := vault.IAMLoginOptions{
		Role:      v.RoleName,
		MountPath: v.AWSAuthPath,
	}

	if _, err := client.Auth.IAM.Login(opts); err != nil {
		return nil, fmt.Errorf("unable to login to Vault: %w", err)
	}

	kvOpts := vault.KV2GetOptions{
		MountPath:     v.SecretMountPath,
		SecretPath:    v.SecretPath,
		UnmarshalInto: &settings,
	}

	if _, err := client.KV2.Get(kvOpts); err != nil {
		return nil, fmt.Errorf("unable to get vault settings from %s/%s: %w", v.SecretMountPath, v.SecretPath, err)
	}

	return &settings, nil
}
