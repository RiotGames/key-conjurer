package keyconjurer

import (
	"fmt"

	"strings"

	"github.com/go-ini/ini"
	homedir "github.com/mitchellh/go-homedir"
)

type awsCli struct {
	creds  *awsCliCredentialsFile
	config *awsCliConfigFile
}

type awsCliCredentialsFile struct {
	*ini.File
	Path string
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

func NewAWSCliEntry(c *Credentials, a *Account) *AWSCliEntry {

	if a.Alias == "" {
		Logger.Warn("Alias is an empty string and profile will be set to default alias")
	}

	a.defaultAlias()

	return &AWSCliEntry{
		profileName: a.Alias,
		keyId:       c.AccessKeyID,
		key:         c.SecretAccessKey,
		token:       c.SessionToken,
	}
}

func getAwsCliCredentialsFile(credsPath string) *awsCliCredentialsFile {
	fullCredsPath, err := homedir.Expand(credsPath)

	if err != nil {
		Logger.Errorln("unable to expand path to aws-cli credentials file")
		Logger.Errorln(err)
		return &awsCliCredentialsFile{}
	}

	Logger.Infof("reading aws-cli creds from %s\n", fullCredsPath)

	creds := &awsCliCredentialsFile{Path: fullCredsPath}

	creds.File, err = ini.Load(fullCredsPath)
	if err != nil {
		Logger.Errorln("could not read aws-cli credentials file")
		Logger.Errorln(err)
		return creds
	}

	return creds
}

func getAwsCliConfigFile(configPath string) *awsCliConfigFile {
	fullConfigPath, err := homedir.Expand(configPath)
	if err != nil {
		Logger.Errorln("unable to expand path to aws-cli config file")
		Logger.Errorln(err)
		return &awsCliConfigFile{}
	}

	Logger.Infof("reading aws-cli config from %v\n", fullConfigPath)

	config := &awsCliConfigFile{Path: fullConfigPath}

	config.File, err = ini.Load(fullConfigPath)
	if err != nil {
		Logger.Errorln("could not read aws-cli config file")
		Logger.Errorln(err)
		return config
	}

	return config
}

func getAwsCliByPath(path string) *awsCli {
	fullPath, err := homedir.Expand(path)
	if err != nil {
		Logger.Errorln("unable to expand path to aws-cli dir")
		Logger.Errorln(err)
		return &awsCli{}
	}

	Logger.Infof("using aws-cli dir %s\n", fullPath)

	var configPath string
	var credsPath string
	if strings.HasSuffix(fullPath, "/") {
		configPath = fmt.Sprintf("%s%s", fullPath, "config")
		credsPath = fmt.Sprintf("%s%s", fullPath, "credentials")
	} else {
		configPath = fmt.Sprintf("%s/%s", fullPath, "config")
		credsPath = fmt.Sprintf("%s/%s", fullPath, "credentials")
	}

	touchFileIfNotExist(configPath)
	touchFileIfNotExist(credsPath)

	return &awsCli{
		creds:  getAwsCliCredentialsFile(credsPath),
		config: getAwsCliConfigFile(configPath),
	}
}

// stub for use if we end up managing config file at some point
// func StubThatDoesNothing(){}
// func saveConfigEntry(alias, region, output string) {}

func (a *awsCli) saveCredentialEntry(entry *AWSCliEntry) {
	var section *ini.Section

	section, err := a.creds.GetSection(entry.profileName)
	if err != nil {
		// create new section
		section, err = a.creds.NewSection(entry.profileName)
		if err != nil {
			Logger.Errorln("error making new aws cli section: ", err)
		}
	}

	if section.HasKey("aws_access_key_id") {
		section.Key("aws_access_key_id").SetValue(entry.keyId)
	} else {
		_, err := section.NewKey("aws_access_key_id", entry.keyId)
		if err != nil {
			Logger.Errorln("error making new aws cli key: ", err)
		}
	}

	if section.HasKey("aws_secret_access_key") {
		section.Key("aws_secret_access_key").SetValue(entry.key)
	} else {
		_, err := section.NewKey("aws_secret_access_key", entry.key)
		if err != nil {
			Logger.Errorln("error making new aws cli key: ", err)
		}
	}

	if section.HasKey("aws_session_token") {
		section.Key("aws_session_token").SetValue(entry.token)
	} else {
		_, err := section.NewKey("aws_session_token", entry.token)
		if err != nil {
			Logger.Errorln("error making new aws cli key: ", err)
		}
	}
}

func SaveAWSCredentialInCLI(awscliPath string, entries ...*AWSCliEntry) error {
	cli := getAwsCliByPath(awscliPath)

	for _, entry := range entries {
		cli.saveCredentialEntry(entry)
	}

	cli.creds.SaveTo(cli.creds.Path)

	return nil
}
