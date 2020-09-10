package keyconjurer

import (
	"log"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	level, err := logrus.ParseLevel("debug")
	if err != nil {
		log.Fatal(err)
	}
	logger.SetLevel(level)

	Logger = logger
}

func TestAwsCliCredsFile(t *testing.T) {
	credsPath := "~/.aws/credentials"
	t.Log("reading in ", credsPath)
	credsFile := getAwsCliCredentialsFile(credsPath)
	for _, section := range credsFile.Sections() {
		t.Log(section.Name(), section.KeyStrings())
	}
}

func TestAwsCliConfigFile(t *testing.T) {
	configPath := "~/.aws/config"
	t.Log("reading in ", configPath)
	configFile := getAwsCliConfigFile(configPath)
	for _, section := range configFile.Sections() {
		t.Log(section.Name(), section.KeyStrings())
	}
}

func TestAwsCliFileNoSlash(t *testing.T) {
	path := "~/.aws"

	awscli := getAwsCliByPath(path)

	for _, section := range awscli.creds.Sections() {
		t.Log(section.Name(), section.Keys())
	}

	for _, section := range awscli.config.Sections() {
		t.Log(section.Name(), section.Keys())
	}
}

func TestAwsCliFileWithSlash(t *testing.T) {
	path := "~/.aws/"

	awscli := getAwsCliByPath(path)

	for _, section := range awscli.creds.Sections() {
		t.Log(section.Name(), section.Keys())
	}

	for _, section := range awscli.config.Sections() {
		t.Log(section.Name(), section.Keys())
	}
}

func TestAddAWSCliEntry(t *testing.T) {
	path := "~/.aws/"

	awscli := getAwsCliByPath(path)

	entry := &AWSCliEntry{
		profileName: "test-profile",
		keyId:       "notanid",
		key:         "notakey",
		token:       "notatoken",
	}

	awscli.saveCredentialEntry(entry)

	assert.Equal(t, true, awscli.creds.Section("test-profile") != nil, "section should have been added above")

	testSection := awscli.creds.Section("test-profile")

	testinikeys := []string{"aws_access_key_id", "aws_secret_access_key", "aws_session_token"}
	testinivals := []string{"notanid", "notakey", "notatoken"}

	for idx, inikey := range testinikeys {
		assert.Equalf(t, true, testSection.HasKey(inikey), "section should have %s field\n")
		key := testSection.Key(inikey)
		assert.Equalf(t, true, key.Value() == testinivals[idx], "field %s should have value %s\n", inikey, testinivals[idx])
	}

	awscli.creds.SaveTo(awscli.creds.Path)

	// retest by reloading into file
	awscli = &awsCli{}
	awscli = getAwsCliByPath(path)

	assert.Equal(t, true, awscli.creds.Section("test-profile") != nil, "section should have been added above")

	testSection = awscli.creds.Section("test-profile")

	for idx, inikey := range testinikeys {
		assert.Equalf(t, true, testSection.HasKey(inikey), "section should have %s field\n")
		key := testSection.Key(inikey)
		assert.Equalf(t, true, key.Value() == testinivals[idx], "field %s should have value %s\n", inikey, testinivals[idx])
	}
}
