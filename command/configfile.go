package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// configFile is the on-disk representation of Config.
//
// It exists so that the configuration file format is not tied to the types
// the rest of the application uses: Config and its related types can be
// reshaped freely without breaking configuration files that are already on
// users' machines. Any field that should be persisted must be added here and
// translated in newConfigFile/toConfig.
type configFile struct {
	Accounts        map[string]configFileAccount `json:"accounts"`
	TTL             uint                         `json:"ttl"`
	TimeRemaining   uint                         `json:"time_remaining"`
	LastUsedAccount *string                      `json:"last_used_account"`
}

type configFileAccount struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Alias          string `json:"alias"`
	MostRecentRole string `json:"most_recent_role"`
}

func newConfigFile(config *Config) configFile {
	file := configFile{
		TTL:             config.TTL,
		TimeRemaining:   config.TimeRemaining,
		LastUsedAccount: config.LastUsedAccount,
		Accounts:        make(map[string]configFileAccount),
	}

	if config.Accounts != nil {
		config.Accounts.ForEach(func(id string, acc Account, _ string) {
			file.Accounts[id] = configFileAccount{
				ID:             acc.ID,
				Name:           acc.Name,
				Alias:          acc.Alias,
				MostRecentRole: acc.MostRecentRole,
			}
		})
	}

	return file
}

func (f configFile) toConfig() Config {
	config := Config{
		TTL:             f.TTL,
		TimeRemaining:   f.TimeRemaining,
		LastUsedAccount: f.LastUsedAccount,
		Accounts:        &accountSet{},
	}

	for id, acc := range f.Accounts {
		config.Accounts.Add(id, Account{
			ID:             acc.ID,
			Name:           acc.Name,
			Alias:          acc.Alias,
			MostRecentRole: acc.MostRecentRole,
		})
	}

	if config.TTL < 1 {
		config.TTL = DefaultTTL
	}

	return config
}

// readConfig populates all member values of config using default values where needed
func readConfig(r io.Reader) (Config, error) {
	var file configFile
	dec := json.NewDecoder(r)
	// If we encounter an end of file, use the default values and don't treat it as an error
	// This also conveniently allows someone to use /dev/null for the config file.
	if err := dec.Decode(&file); err != nil && !errors.Is(err, io.EOF) {
		return Config{}, err
	}

	return file.toConfig(), nil
}

// writeConfig writes the config to the writer provided overwriting the file if it exists
func writeConfig(w io.Writer, config *Config) error {
	enc := json.NewEncoder(w)
	return enc.Encode(newConfigFile(config))
}

func findConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "keyconjurer", "config.json"), nil
}

func loadConfigFromPath(path string) (Config, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		// First run. saveConfig will create the file when the command finishes.
		return configFile{}.toConfig(), nil
	}
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	return readConfig(file)
}

// saveConfigToPath writes the config to a temporary file first and then
// renames it into place, so an interrupted write cannot truncate a
// configuration file that already exists.
func saveConfigToPath(path string, config *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.FileMode(0755)); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), "keyconjurer-config-*")
	if err != nil {
		return fmt.Errorf("unable to create %s reason: %w", path, err)
	}

	if err := writeConfig(tmp, config); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	if err := os.Rename(tmp.Name(), path); err != nil {
		// The rename can fail if something else is holding the configuration
		// file open. We would rather not leave the temporary file behind.
		os.Remove(tmp.Name())
		return err
	}

	return nil
}

func loadConfig() (Config, error) {
	path, err := findConfigPath()
	if err != nil {
		return Config{}, fmt.Errorf("find config path: %s", err)
	}

	return loadConfigFromPath(path)
}

func saveConfig(config *Config) error {
	path, err := findConfigPath()
	if err != nil {
		return fmt.Errorf("find config path: %s", err)
	}

	return saveConfigToPath(path, config)
}
