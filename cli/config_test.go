package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindAccount(t *testing.T) {
	set := accountSet{}
	set.Add("testaccount", Account{ID: "1", Name: "test account"})
	account, ok := set.Resolve("testaccount")
	assert.True(t, ok, "account should exist")
	assert.Equal(t, "1", account.ID, "account id should be 1")
	assert.Equal(t, "test account", account.Name, "account name should be %q", "test account")

	account, ok = set.Resolve("testaccount2")
	assert.False(t, ok, "account shouldn't exist")
}

func TestResolveCanResolveAliases(t *testing.T) {
	set := accountSet{}
	set.Add("1", Account{
		ID:    "1",
		Name:  "testaccount",
		Alias: "totallyacoolalias",
	})

	_, ok := set.Resolve("totallyacoolalias")
	assert.True(t, ok)
}

func TestAccountFuncs(t *testing.T) {
	test := &Account{
		ID:    "12345",
		Name:  "AWS - Test Account",
		Alias: "secondalias",
	}

	assert.True(t, test.IsNameMatch("Test Account"))
	assert.Truef(t, test.IsNameMatch("secondalias"), "Should be able to name match %s with alias %s", "secondalias", test.Alias)
	assert.Equal(t, test.NormalizeName(), "Test Account")
}

func TestUnmarshalJSON(t *testing.T) {
	blob := `{"accounts":{"1":{"id":"1","name":"AWS - name","alias":"name"}},"ttl":1,"time_remaining":0,"creds":"eyJ1c2VybmFtZSI6InVzZXJuYW1lIiwicGFzc3dvcmQiOiJwYXNzd29yZCJ9"}`
	c := Config{}

	assert.NoError(t, json.Unmarshal([]byte(blob), &c))

	acc, ok := c.Accounts.Resolve("name")
	assert.True(t, ok)
	assert.Equal(t, "AWS - name", acc.Name)
	assert.Equal(t, "name", acc.Alias)
	assert.Equal(t, "1", acc.ID)

	assert.Equal(t, uint(0), c.TimeRemaining)
	assert.Equal(t, uint(1), c.TTL)

	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXJuYW1lIiwicGFzc3dvcmQiOiJwYXNzd29yZCJ9", c.Creds)
}

func TestLegacyUnmarshalJSON(t *testing.T) {
	blob := `{"migrated":false,"apps":null,"accounts":{"1":{"id":1,"name":"AWS - name","alias":"name"}},"ttl":1,"time_remaining":0,"creds":"eyJ1c2VybmFtZSI6InVzZXJuYW1lIiwicGFzc3dvcmQiOiJwYXNzd29yZCJ9"}`
	c := Config{}

	assert.NoError(t, json.Unmarshal([]byte(blob), &c))

	acc, ok := c.Accounts.Resolve("name")
	assert.True(t, ok)
	assert.Equal(t, "AWS - name", acc.Name)
	assert.Equal(t, "name", acc.Alias)
	assert.Equal(t, "1", acc.ID)

	assert.Equal(t, uint(0), c.TimeRemaining)
	assert.Equal(t, uint(1), c.TTL)

	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXJuYW1lIiwicGFzc3dvcmQiOiJwYXNzd29yZCJ9", c.Creds)
}

func TestConfigAliasesWork(t *testing.T) {
	cfg := Config{}
	cfg.AddAccount("1234", Account{ID: "1234", Name: "Test Account"})
	_, ok := cfg.FindAccount("alias")
	assert.False(t, ok)

	cfg.Alias("1234", "alias")
	_, ok = cfg.FindAccount("alias")
	assert.True(t, ok)

	cfg.Unalias("alias")
	_, ok = cfg.FindAccount("alias")
	assert.False(t, ok)
}

func TestAliasesPreservedAfterReplaceWith(t *testing.T) {
	cfg := Config{}
	cfg.AddAccount("riot-1", Account{ID: "riot-1", Name: "AWS - riot 1", Alias: "riot-1"})
	cfg.Alias("riot-1", "my-alias")

	_, ok := cfg.FindAccount("riot-1")
	assert.True(t, ok)
	_, ok = cfg.FindAccount("my-alias")
	assert.True(t, ok)

	cfg.UpdateAccounts([]Account{
		{ID: "riot-1", Name: "AWS - riot 1", Alias: ""},
		{ID: "riot-2", Name: "AWS - riot 2", Alias: ""},
	})

	_, ok = cfg.FindAccount("riot-1")
	assert.True(t, ok)
	_, ok = cfg.FindAccount("my-alias")
	assert.True(t, ok)
	_, ok = cfg.FindAccount("riot-2")
	assert.True(t, ok)

	cfg.UpdateAccounts([]Account{
		{ID: "riot-2", Name: "AWS - riot 2", Alias: ""},
	})

	_, ok = cfg.FindAccount("riot-1")
	assert.False(t, ok)
	_, ok = cfg.FindAccount("my-alias")
	assert.False(t, ok)
	_, ok = cfg.FindAccount("riot-2")
	assert.True(t, ok)
}
