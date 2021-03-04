package main

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"

	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/riotgames/key-conjurer/api/core"
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

type accountSet struct {
	accounts map[string]*Account
}

func generateDefaultAlias(name string) string {
	alias := strings.Replace(name, "AWS - ", "", -1)
	alias = strings.Split(alias, " ")[0]
	return strings.ToLower(alias)
}

func (a *accountSet) ForEach(f func(id string, account Account, aliases []string)) {
	for id, acc := range a.accounts {
		f(id, *acc, []string{acc.Alias})
	}
}

// Add adds an account to the set.
func (a *accountSet) Add(id string, account Account) {
	// TODO: This is bad
	if a.accounts == nil {
		a.accounts = make(map[string]*Account)
	}

	a.accounts[id] = &account
}

// Unalias will remove all aliases for an account that matches the given name or given alias.
func (a *accountSet) Unalias(name string) bool {
	for _, acc := range a.accounts {
		if acc.IsNameMatch(name) {
			acc.Alias = ""
			return true
		}
	}

	return false
}

func (a accountSet) Resolve(name string) (*Account, bool) {
	for k, acc := range a.accounts {
		if k == name {
			return acc, true
		}

		if acc.IsNameMatch(name) {
			return acc, true
		}
	}

	return nil, false
}

func (a accountSet) Alias(id, name string) bool {
	entry, ok := a.accounts[id]
	if !ok {
		return false
	}

	entry.Alias = name
	return true
}

func (a *accountSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.accounts)
}

func (a *accountSet) UnmarshalJSON(buf []byte) error {
	var m map[string]struct {
		Name  string
		Alias string
	}

	if err := json.Unmarshal(buf, &m); err != nil {
		return err
	}

	// Now we just need to copy each entry into the set itself
	for id, val := range m {
		a.Add(id, Account{ID: id, Name: val.Name, Alias: val.Alias})
	}

	return nil
}

func (a *accountSet) ReplaceWith(other []Account) {
	if a.accounts == nil {
		a.accounts = make(map[string]*Account)
	}

	m := map[string]struct{}{}
	for _, acc := range other {
		// Preserve the alias if the account ID is the same and it already exists
		if entry, ok := a.accounts[acc.ID]; ok {
			// The name is the only thing that might change.
			entry.Name = acc.Name
		} else {
			a.accounts[acc.ID] = &acc
		}

		m[acc.ID] = struct{}{}
	}

	for k := range a.accounts {
		if _, ok := m[k]; !ok {
			delete(a.accounts, k)
		}
	}
}

func (s accountSet) WriteTable(w io.Writer) {
	tbl := tablewriter.NewWriter(w)
	tbl.SetHeader([]string{"ID", "Name", "Aliases (comma-separated)"})
	s.ForEach(func(id string, acc Account, aliases []string) {
		tbl.Append([]string{id, acc.Name, strings.Join(aliases, ",")})
	})

	tbl.Render()
}

// Config stores all information related to the user
type Config struct {
	Accounts      *accountSet `json:"accounts"`
	Creds         string      `json:"creds"`
	TTL           uint        `json:"ttl"`
	TimeRemaining uint        `json:"time_remaining"`
}

func (c *Config) GetCredentials() (core.Credentials, error) {
	if c.Creds == "" {
		// No credentials have been saved (or they have been cleared recently)
		return core.Credentials{}, ErrNoCredentials
	}

	return core.Credentials{Username: "encrypted", Password: c.Creds}, nil
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
		c.Accounts = &accountSet{}
	}

	if c.TTL < 1 {
		c.TTL = DefaultTTL
	}

	return nil
}

func (c *Config) AddAccount(id string, account Account) {
	if c.Accounts == nil {
		c.Accounts = &accountSet{accounts: make(map[string]*Account)}
	}

	c.Accounts.Add(id, account)
}

func (c *Config) Alias(id, name string) {
	acc, ok := c.Accounts.Resolve(id)
	if !ok {
		return
	}

	acc.Alias = name
}

func (c *Config) Unalias(name string) {
	acc, ok := c.Accounts.Resolve(name)
	if !ok {
		return
	}

	acc.Alias = ""
}

func (c *Config) FindAccount(name string) (Account, bool) {
	if c.Accounts == nil {
		return Account{}, false
	}

	val, ok := c.Accounts.Resolve(name)
	if ok {
		return *val, true
	}

	return Account{}, false
}

func (c *Config) UpdateAccounts(entries []Account) {
	c.Accounts.ReplaceWith(entries)
}

func (c *Config) DumpAccounts(w io.Writer) {
	c.Accounts.WriteTable(w)
}
