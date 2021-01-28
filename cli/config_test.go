package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetTTL(t *testing.T) {
	u := &Config{}

	assert.Equal(t, uint(0), u.TTL, "Config.TTL should be set to default <uint(0)>")
	u.SetTTL(1000)
	assert.Equal(t, uint(1000), u.TTL, "Config.TTL should be set to 1000")
	u.TTL = 1001
	assert.Equal(t, uint(1001), u.TTL, "Config.TTL should be set to 1001")
}

func TestSetTimeRemaining(t *testing.T) {
	u := &Config{}

	assert.Equal(t, uint(0), u.TimeRemaining, "Config.TimeRemaining should be set to default <uint(0)>")
	u.SetTimeRemaining(1000)
	assert.Equal(t, uint(1000), u.TimeRemaining, "Config.TimeRemaining should be set to 1000")
}

func TestFindAccount(t *testing.T) {
	u := &Config{
		Accounts: map[string]*Account{
			"1": {
				ID:    "1",
				Name:  "testaccount",
				Alias: "testaccount",
			},
		},
	}

	account, ok := u.FindAccount("testaccount")
	assert.True(t, ok, "account should exist")
	assert.Equal(t, "1", account.ID, "account id should be 1")
	assert.Equal(t, "testaccount", account.Name, "account name should be \"testaccount\"")

	account, ok = u.FindAccount("testaccount2")
	assert.False(t, ok, "account shouldnt exis't")
	if account != nil {
		t.Fatal("account shouldnt exist and nil should be returned")
	}
}

func TestNewAlias(t *testing.T) {
	u := &Config{
		Accounts: map[string]*Account{
			"1": {
				ID:    "1",
				Name:  "testaccount",
				Alias: "testaccount",
			},
		},
	}

	err := u.NewAlias("testaccount", "ta")
	assert.Equal(t, nil, err, "account should exist and be aliasable")

	foundAccount, ok := u.FindAccount("ta")
	assert.True(t, ok, "account should exist and be found")
	assert.Equal(t, true, foundAccount != nil, "account should exist and not generate error")
}

func TestRemoveAlias(t *testing.T) {
	u := &Config{
		Accounts: map[string]*Account{
			"1": {
				ID:    "1",
				Name:  "testaccount",
				Alias: "totallyacoolalias",
			},
		},
	}

	accountFound, ok := u.FindAccount("totallyacoolalias")
	assert.Equal(t, true, accountFound != nil, "account should exist and be found by alias")
	assert.Equal(t, true, ok, "account should exist and no error generated")

	assert.True(t, u.RemoveAlias("totallyacoolalias"), "should be able to remove alias by looking up account by current alias")

	accountFound, ok = u.FindAccount("totallyacoolalias")
	assert.Equalf(t, true, accountFound == nil, "account shouldn't exist and be found by old alias - <%v>\n", accountFound)
	assert.Equal(t, false, ok, "account shouldnt exist and generate an error")

	accountFound, ok = u.FindAccount("testaccount")
	assert.Equal(t, true, accountFound != nil, "account should exist and be found by default alias")
	assert.Equal(t, true, ok, "account should exist and not generate error")
}

func TestMergeUserData(t *testing.T) {
	u := &Config{}
	toCopy := Config{
		Accounts: map[string]*Account{
			"1": {
				ID:    "1",
				Name:  "testaccount",
				Alias: "totallyacoolalias",
			},
		},
	}

	u.Merge(toCopy)

	accountFound, ok := u.FindAccount("totallyacoolalias")
	assert.Equal(t, true, accountFound != nil, "account should exist and be found by alias")
	assert.True(t, ok, "account should exist and no error generated")
}

func TestAccountFuncs(t *testing.T) {
	test := &Account{
		ID:    "12345",
		Name:  "AWS - Test Account",
		Alias: "",
	}

	test.DefaultAlias()
	assert.Equal(t, test.Alias, "test", "AWS - Test Account should become `test`")
	test.SetAlias("supercooltestalias")
	assert.Equal(t, test.Alias, "supercooltestalias", "Alias should have been set")
	test.SetAlias("secondalias")
	assert.Equal(t, test.Alias, "secondalias", "Alias should have been reassigned")
	assert.Equal(t, test.IsNameMatch("Test Account"), true, "Should be able to name match with normalized name")
	assert.Equalf(t, test.IsNameMatch("secondalias"), true, "Should be able to name match %s with alias %s", "secondalias", test.Alias)
	assert.Equal(t, test.NormalizeName(), "Test Account", true, "Should match normalized name")
}

func TestLegacyUnmarshalJSON(t *testing.T) {
	blob := `{"migrated":false,"apps":null,"accounts":{"1":{"id":"1","name":"AWS - name","alias":"name"}},"ttl":1,"time_remaining":0,"creds":"eyJ1c2VybmFtZSI6InVzZXJuYW1lIiwicGFzc3dvcmQiOiJwYXNzd29yZCJ9"}`
	c := Config{}

	assert.NoError(t, json.Unmarshal([]byte(blob), &c))

	acc, ok := c.FindAccount("name")
	assert.True(t, ok)
	assert.Equal(t, "AWS - name", acc.Name)
	assert.Equal(t, "name", acc.Alias)
	assert.Equal(t, "1", acc.ID)

	assert.Equal(t, uint(0), c.TimeRemaining)
	assert.Equal(t, uint(1), c.TTL)

	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXJuYW1lIiwicGFzc3dvcmQiOiJwYXNzd29yZCJ9", c.Creds)
}

func TestModernUnmarshalJSON(t *testing.T) {

}
