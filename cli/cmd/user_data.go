package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/riotgames/key-conjurer/api/core"
	api "github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/riotgames/key-conjurer/cli/keyconjurer"
)

type accountSet map[string]*keyconjurer.Account

// UserData stores all information related to the user
type UserData struct {
	Accounts      accountSet `json:"accounts"`
	Creds         string     `json:"creds"`
	TTL           uint       `json:"ttl"`
	TimeRemaining uint       `json:"time_remaining"`
}

func (u *UserData) GetCredentials() (core.Credentials, error) {
	if u.Creds == "" {
		// No credentials have been saved (or they have been cleared recently)
		return core.Credentials{}, ErrNoCredentials
	}

	return core.Credentials{Username: "encrypted", Password: u.Creds}, nil
}

func (u *UserData) SetTTL(ttl uint) {
	u.TTL = ttl
}

func (u *UserData) SetTimeRemaining(timeRemaining uint) {
	u.TimeRemaining = timeRemaining
}

func (u *UserData) FindAccount(name string) (*keyconjurer.Account, bool) {
	for _, account := range u.Accounts {
		if account.IsNameMatch(name) {
			return account, true
		}
	}

	return nil, false
}

func (u *UserData) ListAccounts(w io.Writer) error {
	accountTable := tablewriter.NewWriter(w)
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
		if account.IsNameMatch(accountName) {
			account.SetAlias(alias)
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
	account.DefaultAlias()
	return true
}

// Write writes the userData to the file provided overwriting the file if it exists
func (u *UserData) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(u)
}

// Reader populates all member values of userData using default values where needed
func (u *UserData) Read(reader io.Reader) error {
	dec := json.NewDecoder(reader)
	// If we encounter an end of file, use the default values and don't treat it as an error
	// This also conveniently allows someone to use /dev/null for the config file.
	if err := dec.Decode(u); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	if u.TTL < 1 {
		u.SetTTL(keyconjurer.DefaultTTL)
	}

	return nil
}

func (u *UserData) SetDefaults() {
	u.TTL = keyconjurer.DefaultTTL
	u.TimeRemaining = keyconjurer.DefaultTimeRemaining
}

func (u *UserData) UpdateFromServer(r api.GetUserDataPayload) {
	var accounts map[string]*keyconjurer.Account
	for _, app := range r.Apps {
		accounts[app.ID] = &keyconjurer.Account{ID: app.ID, Name: app.Name}
	}

	u.Merge(UserData{Accounts: accounts, Creds: r.EncryptedCredentials})
}

func (u *UserData) mergeAccounts(accounts []core.Application) {
	// This could be improved by simply iterating over the stored accounts, applying aliases to the new accounts and then overwriting the map
	m := map[string]core.Application{}
	for _, acc := range accounts {
		m[acc.ID] = acc
	}

	deleted := []string{}
	for k := range u.Accounts {
		_, ok := m[k]
		if !ok {
			deleted = append(deleted, k)
		}
	}

	for _, acc := range accounts {
		entry, ok := u.Accounts[acc.ID]
		if !ok {
			entry := &keyconjurer.Account{ID: acc.ID, Name: acc.Name}
			entry.DefaultAlias()
			u.Accounts[acc.ID] = entry
		} else {
			entry.Name = acc.Name
			entry.ID = acc.ID
			entry.DefaultAlias()
		}
	}

	for _, k := range deleted {
		delete(u.Accounts, k)
	}
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

	if u.Accounts == nil {
		u.Accounts = map[string]*keyconjurer.Account{}
	}

	for _, app := range toCopy.Accounts {
		acc := &keyconjurer.Account{
			ID:    app.ID,
			Alias: app.Alias,
			Name:  app.Name,
		}
		acc.DefaultAlias()
		u.Accounts[acc.ID] = acc
	}

	for key := range u.Accounts {
		_, ok := toCopy.Accounts[key]
		if !ok {
			delete(u.Accounts, key)
		}
	}
}
