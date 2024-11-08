package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"strings"

	"golang.org/x/oauth2"
)

type TokenSet struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IDToken      string    `json:"id_token"`
	Expiry       time.Time `json:"expiry"`
	TokenType    string    `json:"token_type"`
}

// Token implements oauth2.TokenSource.
func (t TokenSet) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
		TokenType:    t.TokenType,
	}, nil
}

type Account struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Alias          string `json:"alias"`
	MostRecentRole string `json:"most_recent_role"`
}

func (a *Account) NormalizeName() string {
	magicPrefixes := []string{"AWS - ", "Tencent - "}
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

// need support Aws and Tencent
func generateDefaultAlias(name string) string {
	magicPrefixes := []string{"AWS -", "Tencent -", "Tencent Cloud -"}
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
	Tokens          *TokenSet   `json:"tokens"`
	LastUsedAccount *string     `json:"last_used_account"`
}

func (c Config) GetOAuthToken() (*TokenSet, bool) {
	return c.Tokens, c.Tokens != nil
}

func HasTokenExpired(tok *TokenSet) bool {
	if tok == nil {
		return true
	}

	if tok.Expiry.IsZero() {
		return false
	}

	return time.Now().After(tok.Expiry)
}

func (c *Config) SaveOAuthToken(tok *oauth2.Token) error {
	if tok == nil {
		c.Tokens = nil
		return nil
	}

	idToken, _ := tok.Extra("id_token").(string)
	tok2 := TokenSet{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry,
		IDToken:      idToken,
	}

	c.Tokens = &tok2
	return nil
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

func EnsureConfigFileExists(fp string) (io.ReadWriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(fp), os.ModeDir|os.FileMode(0755)); err != nil {
		return nil, err
	}

	return os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}
