package keyconjurer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

// UserData stores all information related to the user
type UserData struct {
	filePath      string
	Migrated      bool              `json:"migrated"`
	Apps          []*App            `json:"apps"`
	Accounts      map[uint]*Account `json:"accounts"`
	Creds         string            `json:"creds"`
	TTL           uint              `json:"ttl"`
	TimeRemaining uint              `json:"time_remaining"`
}

func (u *UserData) SetTTL(ttl uint) {
	u.TTL = ttl
}

func (u *UserData) SetTimeRemaining(timeRemaining uint) {
	u.TimeRemaining = timeRemaining
}

func (u *UserData) FindAccount(accountName string) (*Account, error) {
	for _, account := range u.Accounts {
		if account.isNameMatch(accountName) {
			return account, nil
		}
	}
	return nil, fmt.Errorf("Unable to find account %v", accountName)
}

func (u *UserData) ListAccounts() error {
	accountTable := tablewriter.NewWriter(os.Stdout)
	accountTable.SetHeader([]string{"ID", "Name", "Alias"})

	for _, acc := range u.Accounts {
		accountTable.Append([]string{strconv.FormatUint(uint64(acc.ID), 10), acc.Name, acc.Alias})
	}

	accountTable.Render()

	return nil
}

// NewAlias links an AWS account to a new name for use w/ cli
func (u *UserData) NewAlias(accountName string, alias string) error {
	for _, account := range u.Accounts {
		if account.isNameMatch(accountName) {
			account.setAlias(alias)
			return nil
		}
	}
	return fmt.Errorf("Unable to find account %v and set alias %v", accountName, alias)
}

// RemoveAlias removes the alias associated with the current account
func (u *UserData) RemoveAlias(accountName string) error {
	account, err := u.FindAccount(accountName)
	if err != nil {
		return err
	}

	account.Alias = ""
	account.defaultAlias()
	return nil
}

// Save writes the userData to the file provided overwriting the file if it exists
func (u *UserData) Save() error {
	output, err := json.Marshal(u)
	if err != nil {
		return errors.New("Unable to parse JSON")
	}

	file, err := os.Create(u.filePath)
	if err != nil {
		return fmt.Errorf("Unable to create %s reason: %w", u.filePath, err)
	}
	defer file.Close()
	if _, err := file.Write(output); err != nil {
		return fmt.Errorf("Unable to write %s reason: %w", u.filePath, err)
	}
	return nil
}

// Load populates all member values of userData using default values where needed
func (u *UserData) Load(reader io.Reader) error {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("unable to read %s: %w", u.filePath, err)
	}

	if err := json.Unmarshal(body, u); err != nil {
		return fmt.Errorf("unable to read json in %s: %w", u.filePath, err)
	}

	if u.TTL < 1 {
		u.SetTTL(DefaultTTL)
	}

	if !u.Migrated {
		u.moveAppToAccounts()
	}

	return nil
}

func (u *UserData) SetDefaults() {
	u.TTL = DefaultTTL
	u.TimeRemaining = DefaultTimeRemaining
}

// Prompts the user for the AD credentials and then passes back the list
//  of AWS applications and encrypted creds via the inputed userData
func (u *UserData) promptForADCreds() error {
	username, password, err := getUsernameAndPassword()
	if err != nil {
		return errors.New("Unable to get username or password")
	}

	if err := u.getUserData(username, password); err != nil {
		return errors.New("Unable to login")
	}

	return nil
}

// GetUserData retrieves the list of AWS accounts the user has access too as well as the
//  users encrypted credentials which is passed back via the inputed userData
func (u *UserData) getUserData(username string, password string) error {
	// client and version are build const(vars really)
	data, err := newKeyConjurerUserRequestJSON(Client, Version, username, password)
	if err != nil {
		return err
	}

	responseUserData := UserData{}
	if err := doKeyConjurerAPICall("/get_user_data", data, &responseUserData); err != nil {
		return fmt.Errorf("error calling Key Conjurer /get_user_data api: %w", err)
	}

	u.mergeNewUserData(&responseUserData)
	return nil
}

// Merges Apps (from API) into Accounts since command
// line uses 'accounts' and client code should be easy to understand
func (u *UserData) mergeNewUserData(toCopy *UserData) {
	u.Creds = toCopy.Creds

	if toCopy.TTL != 0 {
		u.TTL = toCopy.TTL
	}

	if toCopy.TimeRemaining != 0 {
		u.TimeRemaining = toCopy.TimeRemaining
	}

	// merge in app and accounts
	//  still use apps but populate accounts
	for _, app := range toCopy.Apps {
		app.defaultAlias()
	}

	if u.Accounts == nil {
		u.Accounts = map[uint]*Account{}
	}

	// since accounts/app are immutable
	// only add if they dont already exist
	for _, app := range toCopy.Apps {
		if _, ok := u.Accounts[app.ID]; !ok {
			acc := &Account{
				ID:    app.ID,
				Alias: app.Alias,
				Name:  app.Name,
			}
			acc.defaultAlias()
			u.Accounts[acc.ID] = acc
		}
	}

	// delete old not currently in accounts
	for key := range u.Accounts {
		keyInList := false
		for _, app := range toCopy.Apps {
			if key == app.ID {
				keyInList = true
				break
			}
		}

		if !keyInList {
			delete(u.Accounts, key)
		}
	}
}

func (u *UserData) moveAppToAccounts() {
	if u.Accounts == nil {
		u.Accounts = map[uint]*Account{}
	}

	for _, app := range u.Apps {
		if _, ok := u.Accounts[app.ID]; !ok {
			u.Accounts[app.ID] = &Account{
				Name:  app.Name,
				ID:    app.ID,
				Alias: app.Alias,
			}
		}
	}

	u.Migrated = true
}
