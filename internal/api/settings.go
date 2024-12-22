package api

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
)

// Settings is used to hold keyconjurer settings
type Settings struct {
	OktaHost  string `json:"oktaHost"`
	OktaToken string `json:"oktaToken"`
}

var SettingsProviders = map[string]SettingsProvider{
	"env": SettingsProviderFunc(RetrieveSettingsFromEnv),
	"vault": VaultRetriever{
		SecretMountPath: os.Getenv("KC_SECRET_MOUNT_PATH"),
		SecretPath:      os.Getenv("KC_SECRET_PATH"),
	},
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
	SecretMountPath string
	SecretPath      string
}

func (v VaultRetriever) FetchSettings(ctx context.Context) (*Settings, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("unable to get Vault client: %w", err)
	}

	kv, err := client.KVv2(v.SecretMountPath).Get(ctx, v.SecretPath)
	if err != nil {
		return nil, err
	}

	settings := Settings{
		OktaHost:  kv.Data["oktaHost"].(string),
		OktaToken: kv.Data["oktaToken"].(string),
	}

	return &settings, nil
}
