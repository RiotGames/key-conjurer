package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
	homedir "github.com/mitchellh/go-homedir"
)

type cloudCli struct {
	creds *cloudCliCredentialsFile
}

// Intentionally missing the `ini` notation sections,keys, and values
//
//	are being handled by the ini library
type CloudCliEntry struct {
	profileName string
	keyId       string
	key         string
	token       string
}

func NewCloudCliEntry(c CloudCredentials, a *Account) *CloudCliEntry {
	name := a.Name
	if a.Alias != "" {
		name = a.Alias
	}

	return &CloudCliEntry{
		profileName: name,
		keyId:       c.AccessKeyID,
		key:         c.SecretAccessKey,
		token:       c.SessionToken,
	}
}

type cloudCliCredentialsFile struct {
	*ini.File
	Path string
}

func touchFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
}

func getCloudCliCredentialsFile(credsPath string) (*cloudCliCredentialsFile, error) {
	path, err := homedir.Expand(credsPath)
	if err != nil {
		return nil, err
	}

	f, err := touchFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var creds cloudCliCredentialsFile
	creds.File, err = ini.Load(f)
	creds.Path = path
	return &creds, err
}

func getCloudCliByPath(path string) (*cloudCli, error) {
	fullPath, err := homedir.Expand(path)
	if err != nil {
		return nil, err
	}

	creds, err := getCloudCliCredentialsFile(filepath.Join(fullPath, "credentials"))
	if err != nil {
		return nil, err
	}

	return &cloudCli{creds: creds}, nil
}

func (a *cloudCli) saveCredentialEntry(entry *CloudCliEntry) error {
	section := a.creds.Section(entry.profileName)
	if strings.Contains(strings.ToLower(a.creds.Path), cloudAws) {
		section.Key("aws_access_key_id").SetValue(entry.keyId)
		section.Key("aws_secret_access_key").SetValue(entry.key)
		section.Key("aws_session_token").SetValue(entry.token)
	} else if strings.Contains(strings.ToLower(a.creds.Path), cloudTencent) {
		section.Key("tencent_access_key_id").SetValue(entry.keyId)
		section.Key("tencent_secret_access_key").SetValue(entry.key)
		section.Key("tencent_session_token").SetValue(entry.token)
	}
	return nil
}

func SaveCloudCredentialInCLI(cloudCliPath string, entries ...*CloudCliEntry) error {
	cli, err := getCloudCliByPath(cloudCliPath)
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
