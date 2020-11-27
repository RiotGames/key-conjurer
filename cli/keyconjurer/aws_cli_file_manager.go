package keyconjurer

import (
	"fmt"
	"io/ioutil"
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
	a.defaultAlias()

	return &AWSCliEntry{
		profileName: a.Alias,
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
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var creds awsCliCredentialsFile
	creds.File, err = ini.Load(b)
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

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var cfg awsCliConfigFile
	cfg.File, err = ini.Load(b)
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

	creds, err := getAwsCliCredentialsFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg, err := getAwsCliConfigFile(credsPath)
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

	section, err := a.creds.GetSection(entry.profileName)
	if err != nil {
		// create new section
		section, err = a.creds.NewSection(entry.profileName)
		return err
	}

	if section.HasKey("aws_access_key_id") {
		section.Key("aws_access_key_id").SetValue(entry.keyId)
	} else {
		_, err := section.NewKey("aws_access_key_id", entry.keyId)
		return err
	}

	if section.HasKey("aws_secret_access_key") {
		section.Key("aws_secret_access_key").SetValue(entry.key)
	} else {
		_, err := section.NewKey("aws_secret_access_key", entry.key)
		return err
	}

	if section.HasKey("aws_session_token") {
		section.Key("aws_session_token").SetValue(entry.token)
	} else {
		_, err := section.NewKey("aws_session_token", entry.token)
		return err
	}

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

	cli.creds.SaveTo(cli.creds.Path)

	return nil
}
