package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	ps "github.com/mitchellh/go-ps"
)

// AWSCredentials are used to store and print out temporary AWS Credentials
// Note: verified that onelogin uses int as ID (so no leading 0's)
// ... but does mean we can have negative user ids
type AWSCredentials struct {
	AccountID       string `json:"AccountId"`
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

// load current ENV credentials is available
func getCredentialsFromENV() *AWSCredentials {
	return &AWSCredentials{
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

func getShellType() string {
	pid := os.Getppid()
	parentProc, _ := ps.FindProcess(pid)
	normalizedName := strings.ToLower(parentProc.Executable())

	if strings.Contains(normalizedName, "powershell") || strings.Contains(normalizedName, "pwsh") {
		return "powershell"
	}
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return normalizedName
}

// PrintCredsForEnv detects the users shell and outputs the credentials for use
//  as environment variables for said shell
func (c AWSCredentials) PrintCredsForEnv() {
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
	fmt.Printf(exportStatement, c.AccessKeyID, c.SecretAccessKey, c.SessionToken, c.SessionToken, c.Expiration, c.AccountID)
}
