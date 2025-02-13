package command

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"strings"
)

type Account struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Alias          string `json:"alias"`
	MostRecentRole string `json:"most_recent_role"`
}

func (a *Account) NormalizeName() string {
	magicPrefixes := []string{"AWS - "}
	name := a.Name
	for _, prefix := range magicPrefixes {
		name = strings.TrimPrefix(name, prefix)
	}

	return name
}

func (a *Account) IsNameMatch(name string) bool {
	// Purposefully not checking the lowercase version of app.Alias
	//  as the user should match the alias provided
	if strings.EqualFold(a.Name, name) {
		return true
	}

	if strings.EqualFold(a.NormalizeName(), name) {
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
	magicPrefixes := []string{"AWS -"}
	for _, prefix := range magicPrefixes {
		name = strings.TrimPrefix(name, prefix)
		name = strings.TrimSpace(name)
	}

	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

func (a *accountSet) ForEach(f func(id string, account Account, alias string)) {
	// Golang does not maintain the order of maps, so we create a slice which is sorted instead.
	var accounts []*Account
	for _, acc := range a.accounts {
		accounts = append(accounts, acc)
	}

	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].Name < accounts[j].Name
	})

	for _, acc := range accounts {
		f(acc.ID, *acc, acc.Alias)
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
	var m map[string]Account

	if err := json.Unmarshal(buf, &m); err != nil {
		return err
	}

	// Now we just need to copy each entry into the set itself
	for id, val := range m {
		a.Add(id, val)
	}

	return nil
}

func (a *accountSet) ReplaceWith(other []Account) {
	if a.accounts == nil {
		a.accounts = make(map[string]*Account)
	}

	m := map[string]struct{}{}
	for _, acc := range other {
		clone := acc
		// Preserve the alias if the account ID is the same and it already exists
		if entry, ok := a.accounts[acc.ID]; ok {
			// The name is the only thing that might change.
			entry.Name = acc.Name
		} else {
			a.accounts[acc.ID] = &clone
		}

		m[acc.ID] = struct{}{}
	}

	for k := range a.accounts {
		if _, ok := m[k]; !ok {
			delete(a.accounts, k)
		}
	}
}

func (a accountSet) WriteTable(w io.Writer, withHeaders bool) {
	tbl := csv.NewWriter(w)
	tbl.Comma = '\t'

	if withHeaders {
		tbl.Write([]string{"id", "name", "alias"})
	}

	a.ForEach(func(id string, acc Account, alias string) {
		tbl.Write([]string{id, acc.Name, alias})
	})

	tbl.Flush()
}

// Config stores all information related to the user
type Config struct {
	Accounts        *accountSet `json:"accounts"`
	TTL             uint        `json:"ttl"`
	TimeRemaining   uint        `json:"time_remaining"`
	LastUsedAccount *string     `json:"last_used_account"`
}

// Encode writes the config to the file provided overwriting the file if it exists
func (c *Config) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(c)
}

// Decode populates all member values of config using default values where needed
func (c *Config) Decode(reader io.Reader) error {
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

func (c *Config) FindAccount(name string) (*Account, bool) {
	if c.Accounts == nil {
		return &Account{}, false
	}

	val, ok := c.Accounts.Resolve(name)
	if ok {
		return val, true
	}

	return &Account{}, false
}

func (c *Config) UpdateAccounts(entries []Account) {
	c.Accounts.ReplaceWith(entries)
}

func (c *Config) DumpAccounts(w io.Writer, withHeaders bool) {
	c.Accounts.WriteTable(w, withHeaders)
}

func ensureConfigFileExists(fp string) (io.ReadWriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(fp), os.ModeDir|os.FileMode(0755)); err != nil {
		return nil, err
	}

	return os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}

func findConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "keyconjurer", "config.json"), nil
}

func loadConfig() (Config, error) {
	var config Config
	path, err := findConfigPath()
	if err != nil {
		return config, fmt.Errorf("find config path: %s", err)
	}

	file, err := ensureConfigFileExists(path)
	if err != nil {
		return config, err
	}

	err = config.Decode(file)
	return config, err
}

func saveConfig(config *Config) error {
	path, err := findConfigPath()
	if err != nil {
		return fmt.Errorf("find config path: %s", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.FileMode(0755)); err != nil {
		return err
	}

	w, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to create %s reason: %w", path, err)
	}
	defer w.Close()

	err = config.Encode(w)
	return err
}
