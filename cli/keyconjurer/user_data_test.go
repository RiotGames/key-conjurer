package keyconjurer

import (
	"testing"

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
		Accounts: map[uint]*Account{
			1: &Account{
				ID:    1,
				Name:  "testaccount",
				Alias: "testaccount",
			},
		},
	}

	account, err := u.FindAccount("testaccount")
	assert.Equal(t, nil, err, "account should exist")
	assert.Equal(t, uint(1), account.ID, "account id should be uint(1)")
	assert.Equal(t, "testaccount", account.Name, "account name should be \"testaccount\"")

	account, err = u.FindAccount("testaccount2")
	assert.Equal(t, false, err == nil, "account shouldnt exist and result in error")
	if account != nil {
		t.Fatal("account shouldnt exist and nil should be returned")
	}
}

func TestNewAlias(t *testing.T) {
	u := &UserData{
		Accounts: map[uint]*Account{
			1: &Account{
				ID:    1,
				Name:  "testaccount",
				Alias: "testaccount",
			},
		},
	}

	err := u.NewAlias("testaccount", "ta")
	assert.Equal(t, nil, err, "account should exist and be aliasable")

	foundAccount, err := u.FindAccount("ta")
	assert.Equal(t, nil, err, "account should exist and be found")
	assert.Equal(t, true, foundAccount != nil, "account should exist and not generate error")
}

func TestRemoveAlias(t *testing.T) {
	u := &UserData{
		Accounts: map[uint]*Account{
			1: &Account{
				ID:    1,
				Name:  "testaccount",
				Alias: "totallyacoolalias",
			},
		},
	}

	accountFound, err := u.FindAccount("totallyacoolalias")
	assert.Equal(t, true, accountFound != nil, "account should exist and be found by alias")
	assert.Equal(t, nil, err, "account should exist and no error generated")

	err = u.RemoveAlias("totallyacoolalias")
	assert.Equal(t, nil, err, "should be able to remove alias by looking up account by current alias")

	accountFound, err = u.FindAccount("totallyacoolalias")
	assert.Equalf(t, true, accountFound == nil, "account shouldn't exist and be found by old alias - <%v>\n", accountFound)
	assert.Equal(t, true, err != nil, "account shouldnt exist and generate an error")

	accountFound, err = u.FindAccount("testaccount")
	assert.Equal(t, true, accountFound != nil, "account should exist and be found by default alias")
	assert.Equal(t, true, err == nil, "account should exist and not generate error")
}

func TestMergeUserData(t *testing.T) {
	t.Log("testing merge from Apps to Accounts")
	u := &UserData{}
	toCopy := &UserData{
		Apps: []*App{
			&App{
				ID:    1,
				Name:  "testaccount",
				Alias: "totallyacoolalias",
			},
		},
	}

	u.mergeNewUserData(toCopy)

	accountFound, err := u.FindAccount("totallyacoolalias")
	assert.Equal(t, true, accountFound != nil, "account should exist and be found by alias")
	assert.Equal(t, nil, err, "account should exist and no error generated")
}

func TestMigrateFunction(t *testing.T) {
	u := &UserData{
		Apps: []*App{
			&App{
				ID:    1,
				Name:  "testaccount",
				Alias: "totallyacoolalias",
			},
		},
	}

	u.moveAppToAccounts()

	accountFound, err := u.FindAccount("totallyacoolalias")
	assert.Equal(t, true, accountFound != nil, "account should exist and be found by alias")
	assert.Equal(t, nil, err, "account should exist and no error generated")
}
