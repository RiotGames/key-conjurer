package command

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	FlagOIDCDomain = "oidc-domain"
	FlagClientID   = "client-id"
	FlagConfigPath = "config"
	FlagQuiet      = "quiet"
	FlagTimeout    = "timeout"
	cloudAws       = "aws"
	cloudTencent   = "tencent"
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
	rootCmd.AddCommand(&switchCmd)
	rootCmd.AddCommand(&aliasCmd)
	rootCmd.AddCommand(&unaliasCmd)
	rootCmd.AddCommand(&rolesCmd)
	rootCmd.SetVersionTemplate("{{.Version}}\n")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "keyconjurer",
	Version: fmt.Sprintf("keyconjurer-%s-%s %s (%s)", runtime.GOOS, runtime.GOARCH, Version, BuildTimestamp),
	Short:   "Retrieve temporary cloud credentials.",
	Long: `KeyConjurer retrieves temporary credentials from Okta with the assistance of an optional API.

To get started run the following commands:
  keyconjurer login
  keyconjurer accounts
  keyconjurer get <accountName>
`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
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

		// We don't care about this being cancelled.
		timeout, _ := cmd.Flags().GetInt(FlagTimeout)
		nextCtx, _ := context.WithTimeout(cmd.Context(), time.Duration(timeout)*time.Second)
		cmd.SetContext(ConfigContext(nextCtx, &config, configPath))
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
		config := ConfigFromCommand(cmd)
		path := ConfigPathFromCommand(cmd)
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
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute(ctx context.Context, args []string) error {
	client := &http.Client{Transport: LogRoundTripper{http.DefaultTransport}}
	ctx = oidc.ClientContext(ctx, client)
	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(ctx)
}
