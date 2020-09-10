package settings

import (
	"log"

	"github.com/riotgames/key-conjurer/api/consts"

	"github.com/sirupsen/logrus"
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
}

type SettingsRetrieverFunc = func(logger *logrus.Entry) *Settings

var SettingsRetrievers = map[string]SettingsRetrieverFunc{}

func NewSettings(logger *logrus.Entry) *Settings {
	logger.Infof("Settings Retriever in Use: %s", consts.SettingsRetrieverSelect)
	return SettingsRetrievers[consts.SettingsRetrieverSelect](logger)
}

func registerRetriever(name string, retrieverFunc SettingsRetrieverFunc) {
	if _, ok := SettingsRetrievers[name]; ok {
		log.Fatalf("Already had registered retriever: %s", name)
	}
	SettingsRetrievers[name] = retrieverFunc
}
