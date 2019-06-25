package keyconjurer

import (
	"encoding/json"
	"keyconjurer-lambda/authenticators"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserDataUnmarshal(t *testing.T) {
	userDataJSON := `
	{
		"devices": [],
		"apps":[],
		"creds": "credstring"
	}
	`

	userData := UserData{}
	err := json.Unmarshal([]byte(userDataJSON), &userData)
	assert.Equal(t, nil, err, "should be able to unmarshal userdata string")

	userDataJSON = `
	{
		"devices": [],
		"apps":[
				{
					"id": 123,
					"name": "test"
				}
			],
		"creds": "credstring"
	}
	`

	userData = UserData{}
	err = json.Unmarshal([]byte(userDataJSON), &userData)
	assert.Equal(t, nil, err, "should be able to unmarshal userdata string")

	assert.Equal(t, 1, len(userData.Apps), "there should be 1 app")
	assert.Equal(t, int64(123), userData.Apps[0].ID(), "id should be int64(123)")
	assert.Equal(t, "test", userData.Apps[0].Name(), "name should be \"test\"")
}

func TestUserDataMarshal(t *testing.T) {
	var userAccount authenticators.Account
	userAccount = UserDataAccount{
		AccountId:   123,
		AccountName: "test",
	}

	_, ok := userAccount.(UserDataAccount)
	assert.Equal(t, true, ok, "UserDataAccount should comply with authenticators.")

	userData := UserData{
		Devices: []Device{},
		Apps:    []authenticators.Account{userAccount},
		Creds:   "test",
	}

	jsonUserData, err := json.Marshal(&userData)
	assert.Equal(t, nil, err, "should be able to marshal userdata")

	t.Log(string(jsonUserData))

	userDataUnmarshalled := UserData{}
	err = json.Unmarshal(jsonUserData, &userDataUnmarshalled)
	assert.Equal(t, nil, err, "should be able to remarshal userdata")

	assert.Equal(t, len(userData.Devices), len(userDataUnmarshalled.Devices), "should have the same number of devices")
	assert.Equal(t, len(userData.Apps), len(userDataUnmarshalled.Apps), "should have the same number of apps")
	assert.Equal(t, userData.Creds, userData.Creds, "should have the same creds")

	appUnmarshalled := userDataUnmarshalled.Apps[0]
	assert.Equal(t, userAccount.ID(), appUnmarshalled.ID(), "should have the same app.ID")
	assert.Equal(t, userAccount.Name(), appUnmarshalled.Name(), "should have the same app.Name")
}
