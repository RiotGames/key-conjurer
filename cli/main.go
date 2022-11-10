package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	keyConjurerRcPath   = pflag.String("keyconjurer-rc-path", "", "Path to KeyConjurer configuration file. If not specified, the home directory will be searched. If specified, will only use this location.")
	ExitCodeOK          = 0
	ExitCodeUsageError  = 1
	ExitCodeConfigError = 2
)

func main() {
	viper.AddConfigPath("$HOME")
	viper.SetConfigType("json")
	viper.SetConfigName(".keyconjurerrc")
	if *keyConjurerRcPath != "" {
		viper.SetConfigFile(*keyConjurerRcPath)
	}

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		// Something went really wrong with Viper.
		// This shouldn't be possible when we're only looking from one file source.
		fmt.Fprintf(os.Stderr, "failed to read configuration: %s", err)
		os.Exit(2)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing configuration: %s", err)
		os.Exit(ExitCodeConfigError)
	}

	err = rootCmd.Execute()
	var usageErr *UsageError
	if err != nil {
		msg := err.Error()
		if errors.As(err, &usageErr) {
			msg = usageErr.Help
		}

		fmt.Fprintln(os.Stderr, msg)
		os.Exit(ExitCodeUsageError)
	}

	// Clean up any configuration changes that were made, if any.
	if viper.ConfigFileUsed() == "" {
		// Viper is able to resolve $HOME when searching, but won't do it when saving, so we do that ourselves.
		path, err := homedir.Expand("~/.keyconjurerrc")
		if err != nil {
			// If this occurs, it means that all the environment variables indicating where the home directory are, are empty.
			// This is pretty unrecoverable and we should have encountered it before we got to this.
			fmt.Fprintf(os.Stderr, "could not save configuration: %s", err)
			os.Exit(ExitCodeConfigError)
		}

		viper.SetConfigFile(path)
	}

	// Once the application is terminating, we need to manually update the configuration changes;
	// Unfortunately, Viper only supports Unmarshalling the configuration into a struct, not the other way around.
	config.UpdateViper(viper.GetViper())

	err = viper.WriteConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to save configuration changes: %s", err)
		os.Exit(ExitCodeConfigError)
	}
}
