package keyconjurer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"
	"github.com/riotgames/key-conjurer/api/core"
	api "github.com/riotgames/key-conjurer/api/keyconjurer"
)

// UserData stores all information related to the user
type UserData struct {
	filePath      string
	Migrated      bool                `json:"migrated"`
	Apps          []*App              `json:"apps"`
	Accounts      map[string]*Account `json:"accounts"`
	Creds         string              `json:"creds"`
	TTL           uint                `json:"ttl"`
	TimeRemaining uint                `json:"time_remaining"`
}

func (u *UserData) GetCredentials() core.Credentials {
	return core.Credentials{Username: "encrypted", Password: u.Creds}
}

func (u *UserData) SetTTL(ttl uint) {
	u.TTL = ttl
}

func (u *UserData) SetTimeRemaining(timeRemaining uint) {
	u.TimeRemaining = timeRemaining
}

func (u *UserData) FindAccount(name string) (*Account, bool) {
	for _, account := range u.Accounts {
		if account.isNameMatch(name) {
			return account, true
		}
	}

	return nil, false
}

func (u *UserData) ListAccounts() error {
	accountTable := tablewriter.NewWriter(os.Stdout)
	accountTable.SetHeader([]string{"ID", "Name", "Alias"})

	for _, acc := range u.Accounts {
		accountTable.Append([]string{acc.ID, acc.Name, acc.Alias})
	}

	accountTable.Render()

	return nil
}

// NewAlias links an AWS account to a new name for use w/ cli
func (u *UserData) NewAlias(accountName string, alias string) error {
	for _, account := range u.Accounts {
		if account.isNameMatch(accountName) {
			account.setAlias(alias)
			return nil
		}
	}
	return fmt.Errorf("Unable to find account %v and set alias %v", accountName, alias)
}

// RemoveAlias removes the alias associated with the current account
func (u *UserData) RemoveAlias(accountName string) bool {
	account, ok := u.FindAccount(accountName)
	if !ok {
		return false
	}

	account.Alias = ""
	account.defaultAlias()
	return true
}

// Save writes the userData to the file provided overwriting the file if it exists
func (u *UserData) Save() error {
	output, err := json.Marshal(u)
	if err != nil {
		return errors.New("Unable to parse JSON")
	}

	file, err := os.Create(u.filePath)
	if err != nil {
		return fmt.Errorf("Unable to create %s reason: %w", u.filePath, err)
	}
	defer file.Close()
	if _, err := file.Write(output); err != nil {
		return fmt.Errorf("Unable to write %s reason: %w", u.filePath, err)
	}
	return nil
}

// SaveToFile saves the UserData to the file system.
func (u *UserData) SaveToFile(fp string) error {
	expanded, err := homedir.Expand(fp)
	if err == nil {
		fp = expanded
	}

	dir := filepath.Dir(expanded)
	if err := os.MkdirAll(dir, os.ModeDir|os.FileMode(0755)); err != nil {
		return err
	}

	u.filePath = fp
	return u.Save()
}

// LoadFromFile loads the UserData from a file.
func (u *UserData) LoadFromFile(fp string) error {
	expanded, err := homedir.Expand(fp)
	if err == nil {
		fp = expanded
	}

	dir := filepath.Dir(expanded)
	if err := os.MkdirAll(dir, os.ModeDir|os.FileMode(0755)); err != nil {
		return err
	}

	file, err := os.OpenFile(expanded, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	u.filePath = fp

	defer file.Close()

	if err := u.Read(file); err != nil {
		return err
	}

	return nil
}

// Load populates all member values of userData using default values where needed
func (u *UserData) Read(reader io.Reader) error {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("unable to read %s: %w", u.filePath, err)
	}

	if err := json.Unmarshal(body, u); err != nil {
		return fmt.Errorf("unable to read json in %s: %w", u.filePath, err)
	}

	if u.TTL < 1 {
		u.SetTTL(DefaultTTL)
	}

	if !u.Migrated {
		u.moveAppToAccounts()
	}

	return nil
}

func (u *UserData) SetDefaults() {
	u.TTL = DefaultTTL
	u.TimeRemaining = DefaultTimeRemaining
}

func (u *UserData) UpdateFromServer(r api.GetUserDataPayload) {
	// This is a bit of a bodge because the server does not actually return a UserData instance but an api.GetUserDataPayload instance.
	// However, there are some shared properties.
	var apps []*App
	for _, app := range r.Apps {
		apps = append(apps, &App{ID: app.ID, Name: app.Name})
	}

	u.Merge(UserData{Apps: apps, Creds: r.EncryptedCredentials})
}

// Merge merges Apps (from API) into Accounts since command line uses 'accounts' and client code should be easy to understand
func (u *UserData) Merge(toCopy UserData) {
	u.Creds = toCopy.Creds

	if toCopy.TTL != 0 {
		u.TTL = toCopy.TTL
	}

	if toCopy.TimeRemaining != 0 {
		u.TimeRemaining = toCopy.TimeRemaining
	}

	// merge in app and accounts
	//  still use apps but populate accounts
	for _, app := range toCopy.Apps {
		app.defaultAlias()
	}

	if u.Accounts == nil {
		u.Accounts = map[string]*Account{}
	}

	// since accounts/app are immutable
	// only add if they dont already exist
	for _, app := range toCopy.Apps {
		if _, ok := u.Accounts[app.ID]; !ok {
			acc := &Account{
				ID:    app.ID,
				Alias: app.Alias,
				Name:  app.Name,
			}
			acc.defaultAlias()
			u.Accounts[acc.ID] = acc
		}
	}

	// delete old not currently in accounts
	for key := range u.Accounts {
		keyInList := false
		for _, app := range toCopy.Apps {
			if key == app.ID {
				keyInList = true
				break
			}
		}

		if !keyInList {
			delete(u.Accounts, key)
		}
	}
}

func (u *UserData) moveAppToAccounts() {
	if u.Accounts == nil {
		u.Accounts = map[string]*Account{}
	}

	for _, app := range u.Apps {
		if _, ok := u.Accounts[app.ID]; !ok {
			u.Accounts[app.ID] = &Account{
				Name:  app.Name,
				ID:    app.ID,
				Alias: app.Alias,
			}
		}
	}

	u.Migrated = true
}
