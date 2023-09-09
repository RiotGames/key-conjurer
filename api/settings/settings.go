package settings

import (
	"fmt"
	"log"
	"os"
)

// Settings is used to hold keyconjurer settings
type Settings struct {
	AwsRegion     string
	AwsKMSKeyID   string `json:"awsKmsKeyId"`
	TencentRegion string
	OktaHost      string `json:"oktaHost" split_words:"true"`
	OktaToken     string `json:"oktaToken" split_words:"true"`
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
		AwsRegion:     os.Getenv("AWS_REGION"),
		TencentRegion: os.Getenv("TENCENT_REGION"),
		AwsKMSKeyID:   os.Getenv("AWS_KMS_KEY_ID"),
		OktaHost:      os.Getenv("OKTA_HOST"),
		OktaToken:     os.Getenv("OKTA_TOKEN"),
	}

	return &s, nil
}

func init() {
	registerRetriever("env", retrieveFromEnv)
	registerRetriever("vault", retrieveFromVault)
	registerRetriever("kms_blob", NewSettingsFromKMSBlob)
}
