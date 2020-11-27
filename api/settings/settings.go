package settings

import (
	"log"

	"github.com/riotgames/key-conjurer/api/consts"
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
	OktaHost               string `json:"oktaHost"`
	OktaToken              string `json:"oktaToken"`
}

type SettingsRetrieverFunc = func() (*Settings, error)

var SettingsRetrievers = map[string]SettingsRetrieverFunc{}

func NewSettings() (*Settings, error) {
	return SettingsRetrievers[consts.SettingsRetrieverSelect]()
}

func registerRetriever(name string, retrieverFunc SettingsRetrieverFunc) {
	if _, ok := SettingsRetrievers[name]; ok {
		log.Fatalf("Already had registered retriever: %s", name)
	}

	SettingsRetrievers[name] = retrieverFunc
}
