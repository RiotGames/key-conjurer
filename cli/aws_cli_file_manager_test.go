package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAwsCliCredsFile(t *testing.T) {
	credsFile, err := getAwsCliCredentialsFile("~/.aws/credentials")
	require.NoError(t, err)

	for _, section := range credsFile.Sections() {
		t.Log(section.Name(), section.KeyStrings())
	}
}

func TestAwsCliConfigFile(t *testing.T) {
	configFile, err := getAwsCliConfigFile("~/.aws/config")
	require.NoError(t, err)

	for _, section := range configFile.Sections() {
		t.Log(section.Name(), section.KeyStrings())
	}
}

func TestAwsCliFileNoSlash(t *testing.T) {
	awscli, err := getAwsCliByPath("~/.aws/")
	require.NoError(t, err)

	for _, section := range awscli.creds.Sections() {
		t.Log(section.Name(), section.Keys())
	}

	for _, section := range awscli.config.Sections() {
		t.Log(section.Name(), section.Keys())
	}
}

func TestAwsCliFileWithSlash(t *testing.T) {
	awscli, err := getAwsCliByPath("~/.aws/")
	require.NoError(t, err)

	for _, section := range awscli.creds.Sections() {
		t.Log(section.Name(), section.Keys())
	}

	for _, section := range awscli.config.Sections() {
		t.Log(section.Name(), section.Keys())
	}
}

func TestAddAWSCliEntry(t *testing.T) {
	awscli, err := getAwsCliByPath("~/.aws/")
	require.NoError(t, err)

	entry := &AWSCliEntry{
		profileName: "test-profile",
		keyId:       "notanid",
		key:         "notakey",
		token:       "notatoken",
	}

	err = awscli.saveCredentialEntry(entry)
	require.NoError(t, err)

	sec := awscli.creds.Section("test-profile")
	require.NotNil(t, sec, "section should have been added above")
	testinikeys := []string{"aws_access_key_id", "aws_secret_access_key", "aws_session_token"}
	testinivals := []string{"notanid", "notakey", "notatoken"}

	for idx, inikey := range testinikeys {
		require.Truef(t, sec.HasKey(inikey), "section should have %s field\n", inikey)
		key := sec.Key(inikey)
		require.Truef(t, key.Value() == testinivals[idx], "field %s should have value %s\n", inikey, testinivals[idx])
	}

	require.NoError(t, awscli.creds.SaveTo(awscli.creds.Path))

	// retest by reloading into file
	awscli, err = getAwsCliByPath("~/.aws/")
	require.NoError(t, err)

	assert.True(t, awscli.creds.Section("test-profile") != nil, "section should have been added above")
	sec = awscli.creds.Section("test-profile")
	for idx, inikey := range testinikeys {
		assert.Truef(t, sec.HasKey(inikey), "section should have %s field\n", inikey)
		key := sec.Key(inikey)
		assert.Truef(t, key.Value() == testinivals[idx], "field %s should have value %s\n", inikey, testinivals[idx])
	}
}
