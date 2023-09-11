package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	FlagOIDCDomain = "oidc-domain"
	FlagClientID   = "client-id"
	FlagConfigPath = "config"
	FlagQuiet      = "quiet"
	FlagTimeout    = "timeout"
)

var (
	buildTimestamp string = BuildDate + " " + BuildTime + " " + BuildTimeZone
	cloudAws              = "aws"
	cloudTencent          = "tencent"
)

func init() {
	rootCmd.PersistentFlags().String(FlagOIDCDomain, OIDCDomain, "The domain name of your OIDC server")
	rootCmd.PersistentFlags().String(FlagClientID, ClientID, "The OAuth2 Client ID for the application registered with your OIDC server")
	rootCmd.PersistentFlags().Int(FlagTimeout, 120, "the amount of time in seconds to wait for keyconjurer to respond")
	rootCmd.PersistentFlags().String(FlagConfigPath, "~/.keyconjurerrc", "path to .keyconjurerrc file")
	rootCmd.PersistentFlags().Bool(FlagQuiet, false, "tells the CLI to be quiet; stdout will not contain human-readable informational messages")
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(&switchCmd)
	rootCmd.AddCommand(&aliasCmd)
	rootCmd.AddCommand(&unaliasCmd)
	rootCmd.AddCommand(&rolesCmd)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     appname,
	Version: fmt.Sprintf("%s %s (%s)", ClientName, Version, buildTimestamp),
	Short:   "Retrieve temporary cloud credentials.",
	Long: `Key Conjurer retrieves temporary credentials from the Key Conjurer API.

To get started run the following commands:
  ` + appname + ` login # You will get prompted for your AD credentials
  ` + appname + ` accounts
  ` + appname + ` get <accountName>
`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var config Config
		// The error of this function call is only non-nil if the flag was not provided or is not a string.
		configPath, _ := cmd.Flags().GetString(FlagConfigPath)
		if expanded, err := homedir.Expand(configPath); err == nil {
			configPath = expanded
		}

		file, err := EnsureConfigFileExists(configPath)
		if err != nil {
			return err
		}

		if err := config.Read(file); err != nil {
			return err
		}

		info := &configInfo{
			Config: &config,
			Path:   configPath,
		}

		// We don't care about this being cancelled.
		timeout, _ := cmd.Flags().GetInt(FlagTimeout)
		nextCtx, _ := context.WithTimeout(cmd.Context(), time.Duration(timeout)*time.Second)
		cmd.SetContext(ConfigContext(nextCtx, info))
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
		config := ConfigFromContext(cmd.Context())
		path := ConfigPathFromContext(cmd.Context())
		if expanded, err := homedir.Expand(path); err == nil {
			path = expanded
		}

		// Do not use EnsureConfigFileExists here!
		// EnsureConfigFileExists opens the file in append mode.
		// If we open the file in append mode, we'll always append to the file. If we open the file in truncate mode before reading from the file, the content will be truncated _before we read from it_, which will cause a users configuration to be discarded every time we run the program.

		if err := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.FileMode(0755)); err != nil {
			return err
		}

		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("unable to create %s reason: %w", path, err)
		}

		defer file.Close()
		return config.Write(file)
	},
}
