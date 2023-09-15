package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
interesting thread on using ENV in unit testing
https://www.reddit.com/r/golang/comments/ar5z3i/how_to_set_env_variables_while_unit_testing/
*/

var envsToUse []string

func init() {
	envsToUse = []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_SESSION_TOKEN",
		"AWSKEY_EXPIRATION",
		"AWSKEY_ACCOUNT",
	}
}

func setEnv(t *testing.T, valid bool) *Account {
	t.Setenv("AWS_ACCESS_KEY_ID", "1234")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "accesskey")
	t.Setenv("AWS_SESSION_TOKEN", "accesstoken")
	t.Setenv("AWSKEY_ACCOUNT", "1234")
	if valid {
		expire := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
		t.Setenv("AWSKEY_EXPIRATION", expire)
	} else {
		expire := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
		t.Setenv("AWSKEY_EXPIRATION", expire)
	}

	return &Account{
		ID:    "1234",
		Name:  "account",
		Alias: "account",
	}
}

func TestGetValidEnvCreds(t *testing.T) {
	account := setEnv(t, true)
	creds := LoadAWSCredentialsFromEnvironment()
	assert.True(t, creds.ValidUntil(account, "aws", 0), "credentials should be valid")
}

func TestGetInvalidEnvCreds(t *testing.T) {
	account := setEnv(t, false)

	// test incorrect time first
	t.Log("testing expired timestamp for key")
	creds := LoadAWSCredentialsFromEnvironment()
	assert.False(t, creds.ValidUntil(account, "aws", 0), "credentials should be invalid due to timestamp")

	account = setEnv(t, true)
	account.ID = ""
	creds = LoadAWSCredentialsFromEnvironment()

	assert.False(t, creds.ValidUntil(account, "aws", 0), "credentials should be invalid due to non-matching id")

	account = setEnv(t, true)
	t.Setenv("AWSKEY_EXPIRATION", "definitely not a timestamp")
	creds = LoadAWSCredentialsFromEnvironment()
	assert.False(t, creds.ValidUntil(account, "aws", 0), "credentials should be invalid due to non-parsable timestamp")
}

func TestTimeWindowEnvCreds(t *testing.T) {
	account := setEnv(t, true)

	t.Log("testing minutes window still within 1hr period for test creds")
	creds := LoadAWSCredentialsFromEnvironment()
	assert.True(t, creds.ValidUntil(account, "aws", 0), "credentials should be valid")
	assert.True(t, creds.ValidUntil(account, "aws", 5), "credentials should be valid")
	assert.True(t, creds.ValidUntil(account, "aws", 30), "credentials should be valid")
	assert.True(t, creds.ValidUntil(account, "aws", 58), "credentials should be valid")

	t.Log("testing minutes window is outside 1hr period for test creds")
	assert.False(t, creds.ValidUntil(account, "aws", 60*time.Minute), "credentials should be valid")
	assert.False(t, creds.ValidUntil(account, "aws", 61*time.Minute), "credentials should be valid")
}
