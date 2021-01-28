package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"

	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/api/keyconjurer"
)

type maybeLegacyID string

func (i *maybeLegacyID) UnmarshalJSON(buf []byte) error {
	var id1 uint64
	var id2 string

	if err := json.Unmarshal(buf, &id1); err == nil {
		*i = maybeLegacyID(strconv.FormatUint(id1, 10))
		return nil
	}

	if err := json.Unmarshal(buf, &id2); err != nil {
		return err
	}

	*i = maybeLegacyID(id2)
	return nil
}

// Account is used to store information related to the AWS OneLogin App/AWS Account
type Account struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

func (a *Account) UnmarshalJSON(buf []byte) error {
	var onDiskRepresentation struct {
		ID    maybeLegacyID `json:"id"`
		Name  string        `json:"name"`
		Alias string        `json:"alias"`
	}

	if err := json.Unmarshal(buf, &onDiskRepresentation); err != nil {
		return err
	}

	a.ID = string(onDiskRepresentation.ID)
	a.Name = onDiskRepresentation.Name
	a.Alias = onDiskRepresentation.Alias
	return nil
}

func (a *Account) NormalizeName() string {
	return strings.Replace(a.Name, "AWS - ", "", -1)
}

func (a *Account) DefaultAlias() {
	if a.Alias == "" {
		alias := strings.Replace(a.Name, "AWS - ", "", -1)
		alias = strings.Split(alias, " ")[0]
		a.Alias = strings.ToLower(alias)
	}
}

func (a *Account) IsNameMatch(name string) bool {
	// Purposefully not checking the lowercase version of app.Alias
	//  as the user should match the alias provided
	if strings.ToLower(a.Name) == strings.ToLower(name) {
		return true
	}

	if strings.ToLower(a.NormalizeName()) == strings.ToLower(name) {
		return true
	}

	if a.Alias == name {
		return true
	}

	return false
}

func (a *Account) SetAlias(alias string) {
	if alias == "" {
		a.DefaultAlias()
	} else {
		a.Alias = alias
	}
}

type accountSet map[string]*Account

// Config stores all information related to the user
type Config struct {
	Accounts      accountSet `json:"accounts"`
	Creds         string     `json:"creds"`
	TTL           uint       `json:"ttl"`
	TimeRemaining uint       `json:"time_remaining"`
}

func (c *Config) GetCredentials() (core.Credentials, error) {
	if c.Creds == "" {
		// No credentials have been saved (or they have been cleared recently)
		return core.Credentials{}, ErrNoCredentials
	}

	return core.Credentials{Username: "encrypted", Password: c.Creds}, nil
}

func (c *Config) SetTTL(ttl uint) {
	c.TTL = ttl
}

func (c *Config) SetTimeRemaining(timeRemaining uint) {
	c.TimeRemaining = timeRemaining
}

func (c *Config) FindAccount(name string) (*Account, bool) {
	for _, account := range c.Accounts {
		if account.IsNameMatch(name) {
			return account, true
		}
	}

	return nil, false
}

func (c *Config) ListAccounts(w io.Writer) error {
	accountTable := tablewriter.NewWriter(w)
	accountTable.SetHeader([]string{"ID", "Name", "Alias"})

	for _, acc := range c.Accounts {
		accountTable.Append([]string{acc.ID, acc.Name, acc.Alias})
	}

	accountTable.Render()

	return nil
}

// NewAlias links an AWS account to a new name for use w/ cli
func (c *Config) NewAlias(accountName string, alias string) error {
	for _, account := range c.Accounts {
		if account.IsNameMatch(accountName) {
			account.SetAlias(alias)
			return nil
		}
	}
	return fmt.Errorf("Unable to find account %v and set alias %v", accountName, alias)
}

// RemoveAlias removes the alias associated with the current account
func (c *Config) RemoveAlias(accountName string) bool {
	account, ok := c.FindAccount(accountName)
	if !ok {
		return false
	}

	account.Alias = ""
	account.DefaultAlias()
	return true
}

// Write writes the config to the file provided overwriting the file if it exists
func (c *Config) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(c)
}

// Reader populates all member values of config using default values where needed
func (c *Config) Read(reader io.Reader) error {
	dec := json.NewDecoder(reader)
	// If we encounter an end of file, use the default values and don't treat it as an error
	// This also conveniently allows someone to use /dev/null for the config file.
	if err := dec.Decode(c); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	if c.Accounts == nil {
		c.Accounts = make(accountSet)
	}

	if c.TTL < 1 {
		c.TTL = DefaultTTL
	}

	return nil
}

func (c *Config) UpdateFromServer(r keyconjurer.GetUserDataPayload) {
	accounts := map[string]*Account{}
	for _, app := range r.Apps {
		accounts[app.ID] = &Account{ID: app.ID, Name: app.Name}
	}

	c.Merge(Config{Accounts: accounts, Creds: r.EncryptedCredentials})
}

func (c *Config) mergeAccounts(accounts []core.Application) {
	// This could be improved by simply iterating over the stored accounts, applying aliases to the new accounts and then overwriting the map
	m := map[string]core.Application{}
	for _, acc := range accounts {
		m[acc.ID] = acc
	}

	deleted := []string{}
	for k := range c.Accounts {
		_, ok := m[k]
		if !ok {
			deleted = append(deleted, k)
		}
	}

	for _, acc := range accounts {
		entry, ok := c.Accounts[acc.ID]
		if !ok {
			entry := &Account{ID: acc.ID, Name: acc.Name}
			entry.DefaultAlias()
			c.Accounts[acc.ID] = entry
		} else {
			entry.Name = acc.Name
			entry.ID = acc.ID
			entry.DefaultAlias()
		}
	}

	for _, k := range deleted {
		delete(c.Accounts, k)
	}
}

// Merge merges Apps (from API) into Accounts since command line uses 'accounts' and client code should be easy to understand
func (c *Config) Merge(toCopy Config) {
	c.Creds = toCopy.Creds

	if toCopy.TTL != 0 {
		c.TTL = toCopy.TTL
	}

	if toCopy.TimeRemaining != 0 {
		c.TimeRemaining = toCopy.TimeRemaining
	}

	if c.Accounts == nil {
		c.Accounts = map[string]*Account{}
	}

	for _, app := range toCopy.Accounts {
		acc := &Account{
			ID:    app.ID,
			Alias: app.Alias,
			Name:  app.Name,
		}
		acc.DefaultAlias()
		c.Accounts[acc.ID] = acc
	}

	for key := range c.Accounts {
		_, ok := toCopy.Accounts[key]
		if !ok {
			delete(c.Accounts, key)
		}
	}
}
