package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
	homedir "github.com/mitchellh/go-homedir"
)

// Intentionally missing the `ini` notation sections,keys, and values are being handled by the ini library
type CloudCliEntry struct {
	profileName string
	keyId       string
	key         string
	token       string
}

func NewCloudCliEntry(c CloudCredentials, a *Account) CloudCliEntry {
	name := a.Name
	if a.Alias != "" {
		name = a.Alias
	}

	return CloudCliEntry{
		profileName: name,
		keyId:       c.AccessKeyID,
		key:         c.SecretAccessKey,
		token:       c.SessionToken,
	}
}

func TouchFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
}

func getCloudCliCredentialsFile(path string) (*ini.File, error) {
	f, err := TouchFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ini.Load(f)
}

func ResolveAWSCredentialsPath(rootPath string) string {
	rootPath = filepath.Join(rootPath, "credentials")
	if fullPath, err := homedir.Expand(rootPath); err == nil {
		return fullPath
	}

	return rootPath
}

func saveCredentialEntry(file *ini.File, entry CloudCliEntry, cloud string) error {
	section := file.Section(entry.profileName)
	if cloud == cloudAws {
		section.Key("aws_access_key_id").SetValue(entry.keyId)
		section.Key("aws_secret_access_key").SetValue(entry.key)
		section.Key("aws_session_token").SetValue(entry.token)
	} else if cloud == cloudTencent {
		section.Key("tencent_access_key_id").SetValue(entry.keyId)
		section.Key("tencent_secret_access_key").SetValue(entry.key)
		section.Key("tencent_session_token").SetValue(entry.token)
	}
	return nil
}

func SaveCloudCredentialInCLI(cloudCliPath string, entry CloudCliEntry) error {
	path := ResolveAWSCredentialsPath(cloudCliPath)
	file, err := getCloudCliCredentialsFile(path)
	if err != nil {
		return err
	}

	cloud := cloudAws
	if strings.Contains(strings.ToLower(path), cloudTencent) {
		cloud = cloudTencent
	}

	if err := saveCredentialEntry(file, entry, cloud); err != nil {
		return err
	}

	return file.SaveTo(path)
}
