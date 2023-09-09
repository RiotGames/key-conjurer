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
)

var (
	//  Config json storage location
	configPath string
	// config is a cache-like datastore for this application. It is loaded at app start-up.
	config         Config
	quiet          bool
	buildTimestamp string = BuildDate + " " + BuildTime + " " + BuildTimeZone
	cloudAws              = "aws"
	cloudTencent          = "tencent"
	timeout        int    = 120
)

func init() {
	rootCmd.PersistentFlags().String(FlagOIDCDomain, OIDCDomain, "The domain name of your OIDC server")
	rootCmd.PersistentFlags().String(FlagClientID, ClientID, "The OAuth2 Client ID for the application registered with your OIDC server")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 120, "the amount of time in seconds to wait for keyconjurer to respond")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "~/.keyconjurerrc", "path to .keyconjurerrc file")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "tells the CLI to be quiet; stdout will not contain human-readable informational messages")
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
		// We don't care about this being cancelled.
		nextCtx, _ := context.WithTimeout(cmd.Context(), time.Duration(timeout)*time.Second)
		cmd.SetContext(nextCtx)

		fp := configPath
		if expanded, err := homedir.Expand(fp); err == nil {
			fp = expanded
		}

		if err := os.MkdirAll(filepath.Dir(fp), os.ModeDir|os.FileMode(0755)); err != nil {
			return err
		}

		file, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return err
		}

		return config.Read(file)
	},
	PersistentPostRunE: func(*cobra.Command, []string) error {
		var fp string
		if expanded, err := homedir.Expand(configPath); err == nil {
			fp = expanded
		}

		dir := filepath.Dir(fp)
		if err := os.MkdirAll(dir, os.ModeDir|os.FileMode(0755)); err != nil {
			return err
		}

		file, err := os.Create(fp)
		if err != nil {
			return fmt.Errorf("unable to create %s reason: %w", fp, err)
		}

		defer file.Close()
		return config.Write(file)
	},
}
