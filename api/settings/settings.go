package settings

import (
	"fmt"
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
	OktaHost               string `json:"oktaHost" split_words:"true"`
	OktaToken              string `json:"oktaToken" split_words:"true"`
}

type retrieverFunc = func() (*Settings, error)

var retrievers = map[string]retrieverFunc{}

func NewSettings() (*Settings, error) {
	prov := os.Getenv("SETTINGS_PROVIDER")
	if prov == "" {
		prov = "env"
	}

	entry, ok := retrievers[prov]
	if !ok {
		return nil, fmt.Errorf("no settings provider with the name %q", prov)
	}

	return entry()
}

func registerRetriever(name string, fn retrieverFunc) {
	if _, ok := retrievers[name]; ok {
		log.Fatalf("Already had registered retriever: %s", name)
	}

	retrievers[name] = fn
}

func retrieveFromEnv() (*Settings, error) {
	s := Settings{
		AwsRegion:              os.Getenv("AWS_REGION"),
		AwsKMSKeyID:            os.Getenv("AWS_KMS_KEY_ID"),
		OneLoginReadUserID:     os.Getenv("ONELOGIN_READ_USER_ID"),
		OneLoginReadUserSecret: os.Getenv("ONELOGIN_READ_USER_SECRET"),
		OneLoginSamlID:         os.Getenv("ONELOGIN_SAML_ID"),
		OneLoginSamlSecret:     os.Getenv("ONELOGIN_SAML_SECRET"),
		OneLoginShard:          os.Getenv("ONELOGIN_SHARD"),
		OneLoginSubdomain:      os.Getenv("ONELOGIN_SUBDOMAIN"),
		OktaHost:               os.Getenv("OKTA_HOST"),
		OktaToken:              os.Getenv("OKTA_TOKEN"),
	}

	return &s, nil
}

func init() {
	registerRetriever("env", retrieveFromEnv)
	registerRetriever("vault", retrieveFromVault)
	registerRetriever("kms_blob", NewSettingsFromKMSBlob)
}
