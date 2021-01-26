package settings

import (
	"log"

	"github.com/kelseyhightower/envconfig"
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

type SettingsRetrieverFunc = func() (*Settings, error)

var SettingsRetrievers = map[string]SettingsRetrieverFunc{}

func NewSettings() (*Settings, error) {
	// TODO: Change back to consts.SettingsRetrieverSelect
	// or at least alter our deployment process so that it uses the appropriate settings retriever
	return SettingsRetrievers["env"]()
}

func registerRetriever(name string, retrieverFunc SettingsRetrieverFunc) {
	if _, ok := SettingsRetrievers[name]; ok {
		log.Fatalf("Already had registered retriever: %s", name)
	}

	SettingsRetrievers[name] = retrieverFunc
}

func retrieveFromEnv() (*Settings, error) {
	var s Settings
	if err := envconfig.Process("keyconjurer", &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func init() {
	registerRetriever("env", retrieveFromEnv)
}
