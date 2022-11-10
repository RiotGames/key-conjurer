package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var keyConjurerRcPath = pflag.String("keyconjurer-rc-path", "", "Path to KeyConjurer configuration file. If not specified, the home directory will be searched. If specified, will only use this location.")

func main() {
	v := viper.New()
	v.AddConfigPath("$HOME")
	v.SetConfigType("json")
	v.SetConfigName(".keyconjurerrc")
	if *keyConjurerRcPath != "" {
		v.SetConfigFile(*keyConjurerRcPath)
	}

	err := v.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
		// Something went really wrong with Viper.
		// This shouldn't be possible when we're only looking from one file source.
		fmt.Fprintf(os.Stderr, "failed to read configuration: %s", err)
		os.Exit(2)
	}

	err = rootCmd.Execute()
	var usageErr *UsageError
	if err != nil {
		msg := err.Error()
		if errors.As(err, &usageErr) {
			msg = usageErr.Help
		}

		fmt.Fprintln(os.Stderr, msg)
		defer os.Exit(1)
	}

	// Clean up any configuration changes that were made, if any.
	if v.ConfigFileUsed() == "" {
		// Viper is able to resolve $HOME when searching, but won't do it when saving, so we do that ourselves.
		path, err := homedir.Expand("~/.keyconjurerrc")
		if err != nil {
			// If this occurs, it means that all the environment variables indicating where the home directory are, are empty.
			// This is pretty unrecoverable and we should have encountered it before we got to this.
			fmt.Fprintf(os.Stderr, "could not save configuration: %s", err)
			os.Exit(2)
		}

		v.SetConfigFile(path)
	}

	err = v.WriteConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to save configuration changes: %s", err)
	}
}
