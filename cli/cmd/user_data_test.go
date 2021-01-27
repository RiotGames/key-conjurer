package cmd

import (
	"testing"

	"github.com/riotgames/key-conjurer/cli/keyconjurer"
	"github.com/stretchr/testify/assert"
)

func TestSetTTL(t *testing.T) {
	u := &UserData{}

	assert.Equal(t, uint(0), u.TTL, "UserData.TTL should be set to default <uint(0)>")
	u.SetTTL(1000)
	assert.Equal(t, uint(1000), u.TTL, "UserData.TTL should be set to 1000")
	u.TTL = 1001
	assert.Equal(t, uint(1001), u.TTL, "UserData.TTL should be set to 1001")
}

func TestSetTimeRemaining(t *testing.T) {
	u := &UserData{}

	assert.Equal(t, uint(0), u.TimeRemaining, "UserData.TimeRemaining should be set to default <uint(0)>")
	u.SetTimeRemaining(1000)
	assert.Equal(t, uint(1000), u.TimeRemaining, "UserData.TimeRemaining should be set to 1000")
}

func TestFindAccount(t *testing.T) {
	u := &UserData{
		Accounts: map[string]*keyconjurer.Account{
			"1": &keyconjurer.Account{
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
	u := &UserData{
		Accounts: map[string]*keyconjurer.Account{
			"1": &keyconjurer.Account{
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
	u := &UserData{
		Accounts: map[string]*keyconjurer.Account{
			"1": &keyconjurer.Account{
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
	u := &UserData{}
	toCopy := UserData{
		Accounts: map[string]*keyconjurer.Account{
			"1": &keyconjurer.Account{
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