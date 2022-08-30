package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-ini/ini"
	homedir "github.com/mitchellh/go-homedir"
)

type awsCli struct {
	creds  *awsCliCredentialsFile
	config *awsCliConfigFile
}

type awsCliConfigFile struct {
	*ini.File
	Path string
}

// Intentionally missing the `ini` notation sections,keys, and values
//  are being handled by the ini library
type AWSCliEntry struct {
	profileName string
	keyId       string
	key         string
	token       string
	region      string
	output      string
}

func NewAWSCliEntry(c *AWSCredentials, a *Account) *AWSCliEntry {
	name := a.Name
	if a.Alias != "" {
		name = a.Alias
	}

	return &AWSCliEntry{
		profileName: name,
		keyId:       c.AccessKeyID,
		key:         c.SecretAccessKey,
		token:       c.SessionToken,
	}
}

type awsCliCredentialsFile struct {
	*ini.File
	Path string
}

func touchFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
}

func getAwsCliCredentialsFile(credsPath string) (*awsCliCredentialsFile, error) {
	path, err := homedir.Expand(credsPath)
	if err != nil {
		return nil, err
	}

	f, err := touchFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var creds awsCliCredentialsFile
	creds.File, err = ini.Load(f)
	creds.Path = path
	return &creds, err
}

func getAwsCliConfigFile(configPath string) (*awsCliConfigFile, error) {
	path, err := homedir.Expand(configPath)
	if err != nil {
		return nil, err
	}

	f, err := touchFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg awsCliConfigFile
	cfg.File, err = ini.Load(f)
	cfg.Path = path
	return &cfg, err
}

func getAwsCliByPath(path string) (*awsCli, error) {
	fullPath, err := homedir.Expand(path)
	if err != nil {
		return nil, err
	}

	var configPath string
	var credsPath string
	if strings.HasSuffix(fullPath, "/") {
		configPath = fmt.Sprintf("%s%s", fullPath, "config")
		credsPath = fmt.Sprintf("%s%s", fullPath, "credentials")
	} else {
		configPath = fmt.Sprintf("%s/%s", fullPath, "config")
		credsPath = fmt.Sprintf("%s/%s", fullPath, "credentials")
	}

	creds, err := getAwsCliCredentialsFile(credsPath)
	if err != nil {
		return nil, err
	}

	cfg, err := getAwsCliConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	return &awsCli{creds: creds, config: cfg}, nil
}

// stub for use if we end up managing config file at some point
// func StubThatDoesNothing(){}
// func saveConfigEntry(alias, region, output string) {}

func (a *awsCli) saveCredentialEntry(entry *AWSCliEntry) error {
	var section *ini.Section
	var err error
	if section, err = a.creds.GetSection(entry.profileName); err != nil {
		if section, err = a.creds.NewSection(entry.profileName); err != nil {
			return err
		}
	}

	section.Key("aws_access_key_id").SetValue(entry.keyId)
	section.Key("aws_secret_access_key").SetValue(entry.key)
	section.Key("aws_session_token").SetValue(entry.token)
	return nil
}

func SaveAWSCredentialInCLI(awscliPath string, entries ...*AWSCliEntry) error {
	cli, err := getAwsCliByPath(awscliPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := cli.saveCredentialEntry(entry); err != nil {
			return err
		}
	}

	return cli.creds.SaveTo(cli.creds.Path)
}
