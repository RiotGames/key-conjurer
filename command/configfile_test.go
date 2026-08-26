package command

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	// This is the format existing installations have on disk.
	// The creds field comes from an older version and should be ignored.
	blob := `{"accounts":{"1":{"id":"1","name":"AWS - name","alias":"name"}},"ttl":1,"time_remaining":0,"creds":"eyJ1c2VybmFtZSI6InVzZXJuYW1lIiwicGFzc3dvcmQiOiJwYXNzd29yZCJ9"}`
	config, err := readConfig(strings.NewReader(blob))
	assert.NoError(t, err)

	acc, ok := config.Accounts.Resolve("name")
	assert.True(t, ok)
	assert.Equal(t, "AWS - name", acc.Name)
	assert.Equal(t, "name", acc.Alias)
	assert.Equal(t, "1", acc.ID)

	assert.Equal(t, uint(0), config.TimeRemaining)
	assert.Equal(t, uint(1), config.TTL)
}

func TestReadConfigLegacy(t *testing.T) {
	blob := `{"migrated":false,"apps":null,"accounts":{"1":{"id":1,"name":"AWS - name","alias":"name"}},"ttl":1,"time_remaining":0,"creds":"eyJ1c2VybmFtZSI6InVzZXJuYW1lIiwicGFzc3dvcmQiOiJwYXNzd29yZCJ9"}`
	_, err := readConfig(strings.NewReader(blob))

	var typeError *json.UnmarshalTypeError
	assert.ErrorAs(t, err, &typeError)
}

func TestReadConfigEmptyFile(t *testing.T) {
	config, err := readConfig(strings.NewReader(""))
	assert.NoError(t, err)
	assert.Equal(t, DefaultTTL, config.TTL)
	assert.NotNil(t, config.Accounts)
}

func TestReadConfigNullAccounts(t *testing.T) {
	// Versions that used os.Create would write null for a config that had
	// no accounts yet.
	blob := `{"accounts":null,"ttl":1,"time_remaining":0,"last_used_account":null}`
	config, err := readConfig(strings.NewReader(blob))
	assert.NoError(t, err)
	assert.NotNil(t, config.Accounts)

	_, ok := config.FindAccount("name")
	assert.False(t, ok)
}

func TestWriteConfig(t *testing.T) {
	var config Config
	config.AddAccount("1", Account{ID: "1", Name: "AWS - name", Alias: "name"})
	config.TTL = 1

	var buf bytes.Buffer
	assert.NoError(t, writeConfig(&buf, &config))

	expected := `{"accounts":{"1":{"id":"1","name":"AWS - name","alias":"name","most_recent_role":""}},"ttl":1,"time_remaining":0,"last_used_account":null}` + "\n"
	assert.Equal(t, expected, buf.String())
}

func TestConfigRoundTrip(t *testing.T) {
	lastUsed := "riot-2"
	config := Config{TTL: 4, TimeRemaining: 30, LastUsedAccount: &lastUsed}
	config.AddAccount("riot-1", Account{ID: "riot-1", Name: "AWS - riot 1", Alias: "riot-1", MostRecentRole: "GL-SuperAdmin"})
	config.AddAccount("riot-2", Account{ID: "riot-2", Name: "AWS - riot 2", Alias: "riot-2"})

	var buf bytes.Buffer
	assert.NoError(t, writeConfig(&buf, &config))

	loaded, err := readConfig(&buf)
	assert.NoError(t, err)

	acc, ok := loaded.FindAccount("riot-1")
	assert.True(t, ok)
	assert.Equal(t, "GL-SuperAdmin", acc.MostRecentRole)

	_, ok = loaded.FindAccount("riot-2")
	assert.True(t, ok)

	assert.Equal(t, uint(4), loaded.TTL)
	assert.Equal(t, uint(30), loaded.TimeRemaining)
	assert.Equal(t, "riot-2", *loaded.LastUsedAccount)
}

func TestSaveConfigToPathReplacesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	first := configFile{TTL: 2}.toConfig()
	assert.NoError(t, saveConfigToPath(path, &first))

	second := configFile{TTL: 8}.toConfig()
	assert.NoError(t, saveConfigToPath(path, &second))

	loaded, err := loadConfigFromPath(path)
	assert.NoError(t, err)
	assert.Equal(t, uint(8), loaded.TTL)

	// The temporary file used during the save should not be left behind.
	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestLoadConfigFromPathMissingFile(t *testing.T) {
	config, err := loadConfigFromPath(filepath.Join(t.TempDir(), "config.json"))
	assert.NoError(t, err)
	assert.Equal(t, DefaultTTL, config.TTL)
	assert.NotNil(t, config.Accounts)
}
