package keyconjurer

import (
	"fmt"
	"os"
	"time"
)

// Credentials are used to store and print out temporary AWS Credentials
// Note: verified that onelogin uses int as ID (so no leading 0's)
// ... but does mean we can have negative user ids
type Credentials struct {
	accountId       uint
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

func GetCredentials(u *UserData, accountName string, ttl uint) (*Credentials, error) {
	// check if account is an account currently assigned
	accountFound, err := u.FindAccount(accountName)
	if err != nil {
		return nil, err
	}

	// check if account asked for is in ENV
	// and still valid
	//
	// on false always build new Credential
	var credentials *Credentials
	switch envCredsValid(accountFound, u.TimeRemaining) {
	case true:
		// use current creds
		credentials = getCredentialsFromENV()
	case false:
		// generate new creds
		var credsTTL uint
		if ttl == 0 {
			credsTTL = u.TTL
		} else {
			credsTTL = ttl
		}
		fmt.Fprintln(os.Stderr, "Sending Duo Push")
		credentials, err = getCredentialsFromKeyConjurer(u.Creds, accountFound, credsTTL)
		if err != nil {
			return nil, err
		}
	}

	credentials.accountId = accountFound.ID

	return credentials, nil
}

// requests a set of temporary credentials for the requested AWS account and returns
//  them via the inputed credentials
func getCredentialsFromKeyConjurer(encryptedAD string, account *Account, ttl uint) (*Credentials, error) {
	data, err := newKeyConjurerCredRequestJSON(Client, Version, "encrypted", encryptedAD, account.ID, ttl)
	if err != nil {
		return nil, err
	}

	responseCredData := Credentials{}
	if err := doKeyConjurerAPICall("/get_aws_creds", data, &responseCredData); err != nil {
		return nil, err
	}

	return &responseCredData, nil
}

// load current ENV credentials is available
func getCredentialsFromENV() *Credentials {
	return &Credentials{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
		Expiration:      os.Getenv("AWSKEY_EXPIRATION"),
	}
}

func envCredsValid(account *Account, minutesTimeWindow uint) bool {
	// if no env var AWSKEY_ACCOUNT or they dont match
	// generate new creds
	currentAccount, ok := os.LookupEnv("AWSKEY_ACCOUNT")
	if !ok || currentAccount != fmt.Sprint(account.ID) {
		return false
	}

	// if no env var AWSKEY_EXPIRATION
	// generate new creds
	currentExpiration, ok := os.LookupEnv("AWSKEY_EXPIRATION")
	if !ok {
		return false
	}

	// if expiration cant be parsed
	// generate new creds
	expiration, err := time.Parse(time.RFC3339, currentExpiration)
	if err != nil {
		return false
	}

	// use creds if they havent expired or generate new one is they have
	// also take into account a time window in which the creds must still be valid
	// example: the creds must still be valid in now + 5m
	return expiration.After(time.Now().Add(time.Minute * time.Duration(minutesTimeWindow)))
}

/*
Help Funcs
*/

// PrintCredsForEnv detects the users shell and outputs the credentials for use
//  as environment variables for said shell
func (c Credentials) PrintCredsForEnv() {
	exportStatement := ""
	switch getShellType() {
	case "powershell":
		exportStatement = `$Env:AWS_ACCESS_KEY_ID = "%v"
$Env:AWS_SECRET_ACCESS_KEY = "%v"
$Env:AWS_SESSION_TOKEN = "%v"
$Env:AWS_SECURITY_TOKEN = "%v"
$Env:TF_VAR_access_key = $Env:AWS_ACCESS_KEY_ID
$Env:TF_VAR_secret_key = $Env:AWS_SECRET_ACCESS_KEY
$Env:TF_VAR_token = $Env:AWS_SESSION_TOKEN
$Env:AWSKEY_EXPIRATION = "%v"
$Env:AWSKEY_ACCOUNT = "%v"
`
	case "cmd":
		exportStatement = `SET AWS_ACCESS_KEY_ID=%v
SET AWS_SECRET_ACCESS_KEY=%v
SET AWS_SESSION_TOKEN=%v
SET AWS_SECURITY_TOKEN=%v
SET TF_VAR_access_key=%%AWS_ACCESS_KEY_ID%%
SET TF_VAR_secret_key=%%AWS_SECRET_ACCESS_KEY%%
SET TF_VAR_token=%%AWS_SESSION_TOKEN%%
SET AWSKEY_EXPIRATION=%v
SET AWSKEY_ACCOUNT=%v
`
	case "bash":
		fallthrough
	default:
		exportStatement = `export AWS_ACCESS_KEY_ID=%v
export AWS_SECRET_ACCESS_KEY=%v
export AWS_SESSION_TOKEN=%v
export AWS_SECURITY_TOKEN=%v
export TF_VAR_access_key=$AWS_ACCESS_KEY_ID
export TF_VAR_secret_key=$AWS_SECRET_ACCESS_KEY
export TF_VAR_token=$AWS_SESSION_TOKEN
export AWSKEY_EXPIRATION=%v
export AWSKEY_ACCOUNT=%v
`
	}
	fmt.Printf(exportStatement, c.AccessKeyID, c.SecretAccessKey, c.SessionToken,
		c.SessionToken, c.Expiration, c.accountId)
}
