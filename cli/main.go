package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

func main() {
	// This var block is defined inside main() to prevent polluting the package environment.
	var (
		keyConjurerRcPath = pflag.String("keyconjurer-rc-path", "", "Path to KeyConjurer configuration file. If not specified, the home directory will be searched. If specified, will only use this location.")
	)

	v := viper.New()
	v.AddConfigPath("$HOME")
	v.AddConfigPath(".")
	v.SetConfigName(".keyconjurerrc")
	v.SetConfigType("json")
	if keyConjurerRcPath != nil {
		v.SetConfigFile(*keyConjurerRcPath)
	}

	err := v.ReadInConfig()
	var opts slog.HandlerOptions
	if _, ok := os.LookupEnv("DEBUG"); ok {
		opts.Level = slog.DebugLevel
	}

	handler := opts.NewTextHandler(os.Stdout)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	if err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			slog.Debug("no configuration file found")
		}
	}

	err = rootCmd.Execute()
	if err == nil {
		return
	}

	var usageErr *UsageError
	if errors.As(err, &usageErr) {
		fmt.Fprintln(os.Stderr, usageErr.Help)
		os.Exit(1)
		return
	}

	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
