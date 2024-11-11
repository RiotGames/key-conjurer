package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/aws"
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

func (v VaultRetriever) FetchSettings(ctx context.Context) (*Settings, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("unable to get Vault client: %w", err)
	}

	auth, err := aws.NewAWSAuth(aws.WithIAMAuth(), aws.WithMountPath(v.AWSAuthPath), aws.WithRole(v.RoleName))
	if err != nil {
		return nil, err
	}

	_, err = client.Auth().Login(ctx, auth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to Vault: %w", err)
	}

	kv, err := client.KVv2(v.SecretMountPath).Get(ctx, v.SecretPath)
	if err != nil {
		return nil, err
	}

	var settings Settings

	jsonBlob, ok := kv.Data["data"].(string)
	if !ok {
		return nil, fmt.Errorf("settings stored in Vault path %s are not a JSON string", fmt.Sprintf("%s/%s", v.SecretMountPath, v.SecretPath))
	}

	err = json.Unmarshal([]byte(jsonBlob), &settings)
	return &settings, err
}
