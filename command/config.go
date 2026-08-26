package command

import (
	"encoding/csv"
	"io"
	"sort"
	"strings"
)

type Account struct {
	ID             string
	Name           string
	Alias          string
	MostRecentRole string
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

func (a *accountSet) ensureInit() {
	if a.accounts == nil {
		a.accounts = make(map[string]*Account)
	}
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
	a.ensureInit()
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

func (a *accountSet) ReplaceWith(other []Account) {
	a.ensureInit()

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
	Accounts        *accountSet
	TTL             uint
	TimeRemaining   uint
	LastUsedAccount *string
}

func (c *Config) AddAccount(id string, account Account) {
	if c.Accounts == nil {
		c.Accounts = &accountSet{}
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
