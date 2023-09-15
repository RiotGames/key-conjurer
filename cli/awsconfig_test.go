package main

import (
	"bytes"
	"testing"

	"github.com/go-ini/ini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddAWSCliEntry(t *testing.T) {
	file, err := ini.Load(bytes.NewReader([]byte{}))
	require.NoError(t, err)

	entry := CloudCliEntry{
		profileName: "test-profile",
		keyId:       "notanid",
		key:         "notakey",
		token:       "notatoken",
	}

	require.NoError(t, saveCredentialEntry(file, entry, cloudAws))

	sec := file.Section("test-profile")
	require.NotNil(t, sec, "section should have been added above")
	testinikeys := []string{"aws_access_key_id", "aws_secret_access_key", "aws_session_token"}
	testinivals := []string{"notanid", "notakey", "notatoken"}

	for idx, inikey := range testinikeys {
		require.Truef(t, sec.HasKey(inikey), "section should have %s field\n", inikey)
		key := sec.Key(inikey)
		require.Truef(t, key.Value() == testinivals[idx], "field %s should have value %s\n", inikey, testinivals[idx])
	}

	var buf bytes.Buffer
	_, err = file.WriteTo(&buf)
	require.NoError(t, err)

	file2, err := ini.Load(&buf)
	require.NoError(t, err)

	assert.True(t, file2.Section("test-profile") != nil, "section should have been added above")
	sec = file2.Section("test-profile")
	for idx, inikey := range testinikeys {
		assert.Truef(t, sec.HasKey(inikey), "section should have %s field\n", inikey)
		key := sec.Key(inikey)
		assert.Truef(t, key.Value() == testinivals[idx], "field %s should have value %s\n", inikey, testinivals[idx])
	}
}
