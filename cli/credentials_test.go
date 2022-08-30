package main

import (
	"os"
	"strings"
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

func resetEnv(t *testing.T, env []string) {
	currentEnv := map[string]string{}
	for _, kvstring := range env {
		kv := strings.Split(kvstring, "=")
		currentEnv[kv[0]] = kv[1]
	}

	for _, resetVar := range envsToUse {
		t.Log("clearing env var: ", resetVar)
		os.Unsetenv(resetVar)
		if value, ok := currentEnv[resetVar]; ok {
			os.Setenv(resetVar, value)
		}
	}

}

func setEnv(t *testing.T, valid bool) *Account {
	if err := os.Setenv("AWS_ACCESS_KEY_ID", "1234"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("AWS_SECRET_ACCESS_KEY", "accesskey"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("AWS_SESSION_TOKEN", "accesstoken"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("AWSKEY_ACCOUNT", "1234"); err != nil {
		t.Fatal(err)
	}

	if valid {
		expire := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
		if err := os.Setenv("AWSKEY_EXPIRATION", expire); err != nil {
			t.Fatal(err)
		}

	} else {
		expire := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
		if err := os.Setenv("AWSKEY_EXPIRATION", expire); err != nil {
			t.Fatal(err)
		}
	}

	return &Account{
		ID:    "1234",
		Name:  "account",
		Alias: "account",
	}
}

func TestResetENV(t *testing.T) {
	if err := os.Setenv("AWSKEY_ACCOUNT", "1234"); err != nil {
		t.Fatal(err)
	}

	envToReset := os.Environ()
	if err := os.Unsetenv("AWSKEY_ACCOUNT"); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv("AWSKEY_ACCOUNT", "5678"); err != nil {
		t.Fatal(err)
	}

	assert.Equalf(t, "5678", os.Getenv("AWSKEY_ACCOUNT"), "Env var should be 1234 but is: %s", os.Getenv("AWSKEY_ACCOUNT"))

	resetEnv(t, envToReset)

	assert.Equalf(t, "1234", os.Getenv("AWSKEY_ACCOUNT"), "Env var should be 1234 but is: %s", os.Getenv("AWSKEY_ACCOUNT"))
}

func TestGetValidEnvCreds(t *testing.T) {
	defer resetEnv(t, os.Environ())
	account := setEnv(t, true)

	var creds AWSCredentials
	creds.LoadFromEnv()
	assert.True(t, creds.ValidUntil(*account, 0), "credentials should be valid")
}

func TestGetInvalidEnvCreds(t *testing.T) {
	defer resetEnv(t, os.Environ())
	account := setEnv(t, false)

	// test incorrect time first
	t.Log("testing expired timestamp for key")
	var creds AWSCredentials
	creds.LoadFromEnv()
	assert.False(t, creds.ValidUntil(*account, 0), "credentials should be invalid due to timestamp")

	account = setEnv(t, true)
	account.ID = ""
	creds.LoadFromEnv()
	assert.False(t, creds.ValidUntil(*account, 0), "credentials should be invalid due to non-matching id")

	account = setEnv(t, true)
	creds.LoadFromEnv()
	if err := os.Setenv("AWSKEY_EXPIRATION", "definitely not a timestamp"); err != nil {
		t.Fatal("unable to reset timestamp to be unparsable")
	}

	assert.False(t, creds.ValidUntil(*account, 0), "credentials should be invalid due to non-parsable timestamp")
}

func TestTimeWindowEnvCreds(t *testing.T) {
	defer resetEnv(t, os.Environ())
	account := setEnv(t, true)

	t.Log("testing minutes window still within 1hr period for test creds")
	var creds AWSCredentials
	creds.LoadFromEnv()
	assert.True(t, creds.ValidUntil(*account, 0), "credentials should be valid")
	assert.True(t, creds.ValidUntil(*account, 5), "credentials should be valid")
	assert.True(t, creds.ValidUntil(*account, 30), "credentials should be valid")
	assert.True(t, creds.ValidUntil(*account, 58), "credentials should be valid")

	t.Log("testing minutes window is outside 1hr period for test creds")
	assert.False(t, creds.ValidUntil(*account, 60*time.Minute), "credentials should be valid")
	assert.False(t, creds.ValidUntil(*account, 61*time.Minute), "credentials should be valid")
}
