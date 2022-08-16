package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	ps "github.com/mitchellh/go-ps"
)

type ShellType = string

const (
	shellTypePowershell ShellType = "powershell"
	shellTypeBash       ShellType = "bash"
	shellTypeBasic      ShellType = "basic"
	shellTypeInfer      ShellType = "infer"
)

func getShellType() ShellType {
	pid := os.Getppid()
	parentProc, _ := ps.FindProcess(pid)
	normalizedName := strings.ToLower(parentProc.Executable())

	if strings.Contains(normalizedName, "powershell") || strings.Contains(normalizedName, "pwsh") {
		return shellTypePowershell
	}

	if runtime.GOOS == "windows" {
		return shellTypeBasic
	}

	return shellTypeBash
}

// CloudCredentials are used to store and print out temporary AWS Credentials
// Note: verified that onelogin uses int as ID (so no leading 0's)
// ... but does mean we can have negative user ids
type CloudCredentials struct {
	AccountID       string `json:"AccountId"`
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

func (c *CloudCredentials) LoadFromEnv(cloudFlag string) {
	if cloudFlag == cloudTencent {
		c.AccountID = os.Getenv("TENCENTKEY_ACCOUNT")
		c.AccessKeyID = os.Getenv("TENCENTCLOUD_SECRET_ID")
		c.SecretAccessKey = os.Getenv("TENCENTCLOUD_SECRET_KEY")
		c.SessionToken = os.Getenv("TENCENTCLOUD_TOKEN")
		c.Expiration = os.Getenv("TENCENTKEY_EXPIRATION")
	} else if cloudFlag == cloudAws {
		c.AccountID = os.Getenv("AWSKEY_ACCOUNT")
		c.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
		c.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
		c.SessionToken = os.Getenv("AWS_SESSION_TOKEN")
		c.Expiration = os.Getenv("AWSKEY_EXPIRATION")
	}
}

func (c *CloudCredentials) ValidUntil(account Account, cloudFlag string, dur time.Duration) bool {
	currentAccount, ok := os.LookupEnv("AWSKEY_ACCOUNT")
	if cloudFlag == cloudTencent {
		currentAccount, ok = os.LookupEnv("TENCENTKEY_ACCOUNT")
	}

	if !ok || currentAccount != account.ID {
		return false
	}

	currentExpiration, ok := os.LookupEnv("AWSKEY_EXPIRATION")
	if cloudFlag == cloudTencent {
		currentExpiration, ok = os.LookupEnv("TENCENTKEY_EXPIRATION")
	}

	if !ok {
		return false
	}

	expiration, err := time.Parse(time.RFC3339, currentExpiration)
	if err != nil {
		return false
	}

	return expiration.After(time.Now().Add(dur))
}

const (
	aws_shellTypePowershell = `$Env:AWS_ACCESS_KEY_ID = "%v"
$Env:AWS_SECRET_ACCESS_KEY = "%v"
$Env:AWS_SESSION_TOKEN = "%v"
$Env:AWS_SECURITY_TOKEN = "%v"
$Env:TF_VAR_access_key = $Env:AWS_ACCESS_KEY_ID
$Env:TF_VAR_secret_key = $Env:AWS_SECRET_ACCESS_KEY
$Env:TF_VAR_token = $Env:AWS_SESSION_TOKEN
$Env:AWSKEY_EXPIRATION = "%v"
$Env:AWSKEY_ACCOUNT = "%v"
`
	tencent_shellTypePowershell = `$Env:TENCENTCLOUD_SECRET_ID = "%v"
$Env:TENCENTCLOUD_SECRET_KEY = "%v"
$Env:TENCENTCLOUD_TOKEN = "%v"
$Env:TENCENTCLOUD_SECURITY_TOKEN = "%v"
$Env:TF_VAR_access_key = $Env:TENCENTCLOUD_SECRET_ID
$Env:TF_VAR_secret_key = $Env:TENCENTCLOUD_SECRET_KEY
$Env:TF_VAR_token = $Env:TENCENTCLOUD_TOKEN
$Env:TENCENT_KEY_EXPIRATION = "%v"
$Env:TENCENT_KEY_ACCOUNT = "%v"
`
	aws_shellTypeBasic = `SET AWS_ACCESS_KEY_ID=%v
SET AWS_SECRET_ACCESS_KEY=%v
SET AWS_SESSION_TOKEN=%v
SET AWS_SECURITY_TOKEN=%v
SET TF_VAR_access_key=%%AWS_ACCESS_KEY_ID%%
SET TF_VAR_secret_key=%%AWS_SECRET_ACCESS_KEY%%
SET TF_VAR_token=%%AWS_SESSION_TOKEN%%
SET AWSKEY_EXPIRATION=%v
SET AWSKEY_ACCOUNT=%v
`
	tencent_shellTypeBasic = `SET TENCENTCLOUD_SECRET_ID=%v
SET TENCENTCLOUD_SECRET_KEY=%v
SET TENCENTCLOUD_TOKEN=%v
SET TENCENTCLOUD_SECURITY_TOKEN=%v
SET TF_VAR_access_key=%%TENCENTCLOUD_SECRET_ID%%
SET TF_VAR_secret_key=%%TENCENTCLOUD_SECRET_KEY%%
SET TF_VAR_token=%%TENCENTCLOUD_TOKEN%%
SET TENCENTKEY_EXPIRATION=%v
SET TENCENTKEY_ACCOUNT=%v`
	aws_shellTypeBash = `export AWS_ACCESS_KEY_ID=%v
export AWS_SECRET_ACCESS_KEY=%v
export AWS_SESSION_TOKEN=%v
export AWS_SECURITY_TOKEN=%v
export TF_VAR_access_key=$AWS_ACCESS_KEY_ID
export TF_VAR_secret_key=$AWS_SECRET_ACCESS_KEY
export TF_VAR_token=$AWS_SESSION_TOKEN
export AWSKEY_EXPIRATION=%v
export AWSKEY_ACCOUNT=%v
`
	tencent_shellTypeBash = `export TENCENTCLOUD_SECRET_ID=%v
export TENCENTCLOUD_SECRET_KEY=%v
export TENCENTCLOUD_TOKEN=%v
export TENCENT_SECURITY_TOKEN=%v
export TF_VAR_access_key=$TENCENTCLOUD_SECRET_ID
export TF_VAR_secret_key=$TENCENTCLOUD_SECRET_KEY
export TF_VAR_token=$TENCENTCLOUD_TOKEN
export TENCENTKEY_EXPIRATION=%v
export TENCENTKEY_ACCOUNT=%v
`
)

func (c CloudCredentials) WriteFormat(w io.Writer, format ShellType, cloudFlag string) (int, error) {
	var str string
	if format == shellTypeInfer {
		format = getShellType()
	}

	switch format {
	case shellTypePowershell:
		str = aws_shellTypePowershell
		if cloudFlag == cloudTencent {
			str = tencent_shellTypePowershell
		}
	case shellTypeBasic:
		str = aws_shellTypeBasic
		if cloudFlag == cloudTencent {
			str = tencent_shellTypeBasic
		}
	case shellTypeBash:
		str = aws_shellTypeBash
		if cloudFlag == cloudTencent {
			str = tencent_shellTypeBash
		}
	}

	return fmt.Fprintf(w, str, c.AccessKeyID, c.SecretAccessKey, c.SessionToken, c.SessionToken, c.Expiration, c.AccountID)
}
