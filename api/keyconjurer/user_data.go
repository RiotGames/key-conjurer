package keyconjurer

import (
	"encoding/json"
	"keyconjurer-lambda/authenticators"
)

// UserData holds a users devices, apps, and encrypted creds
type UserData struct {
	// Devices is an artifact from the past.  We should be
	//  able to pull this out of our code after April 1, 2019
	Devices []Device                 `json:"devices"`
	Apps    []authenticators.Account `json:"apps"`
	Creds   string                   `json:"creds"`
}

// SetCreds is used to update the users encrypted Creds for
//  the response
func (u *UserData) SetCreds(creds string) {
	u.Creds = creds
}

func (u *UserData) MarshalJSON() ([]byte, error) {
	marshalMap := map[string]*json.RawMessage{}

	var err error
	var jsonDevices json.RawMessage
	jsonDevices, err = json.Marshal(&u.Devices)
	if err != nil {
		return []byte{}, nil
	}

	marshalMap["devices"] = &jsonDevices

	apps := []UserDataAccount{}
	for _, account := range u.Apps {
		apps = append(apps, UserDataAccount{AccountId: account.ID(), AccountName: account.Name()})
	}

	var jsonApps json.RawMessage
	jsonApps, err = json.Marshal(&apps)
	if err != nil {
		return []byte{}, nil
	}

	marshalMap["apps"] = &jsonApps

	var jsonCreds json.RawMessage
	jsonCreds, err = json.Marshal(&u.Creds)
	if err != nil {
		return []byte{}, nil
	}

	marshalMap["creds"] = &jsonCreds

	return json.Marshal(&marshalMap)
}

func (u *UserData) UnmarshalJSON(data []byte) error {
	var objmap map[string]*json.RawMessage

	err := json.Unmarshal(data, &objmap)
	if err != nil {
		return nil
	}

	var devices []Device
	err = json.Unmarshal(*objmap["devices"], &devices)
	if err != nil {
		return nil
	}

	var apps []UserDataAccount
	err = json.Unmarshal(*objmap["apps"], &apps)
	if err != nil {
		return nil
	}

	var creds string
	err = json.Unmarshal(*objmap["creds"], &creds)

	authAccounts := make([]authenticators.Account, len(apps))
	for index, app := range apps {
		authAccounts[index] = app
	}

	u.Devices = devices
	u.Apps = authAccounts
	u.Creds = creds

	return nil
}

type UserDataAccount struct {
	AccountId   int64  `json:"id"`
	AccountName string `json:"name"`
}

func (uda UserDataAccount) ID() int64 {
	return uda.AccountId
}

func (uda UserDataAccount) Name() string {
	return uda.AccountName
}

// Device is an artifact from older times we should be able to
//  remove this after April 1 , 2019
type Device struct {
}
