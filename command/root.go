package command

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/spf13/cobra"
)

var (
	FlagOIDCDomain = "oidc-domain"
	FlagClientID   = "client-id"
	FlagQuiet      = "quiet"
	FlagTimeout    = "timeout"
)

func init() {
	rootCmd.PersistentFlags().String(FlagOIDCDomain, OIDCDomain, "The domain name of your OIDC server")
	rootCmd.PersistentFlags().String(FlagClientID, ClientID, "The OAuth2 Client ID for the application registered with your OIDC server")
	rootCmd.PersistentFlags().Int(FlagTimeout, 120, "the amount of time in seconds to wait for keyconjurer to respond")
	rootCmd.PersistentFlags().Bool(FlagQuiet, false, "tells the CLI to be quiet; stdout will not contain human-readable informational messages")
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(&switchCmd)
	rootCmd.AddCommand(&aliasCmd)
	rootCmd.AddCommand(&unaliasCmd)
	rootCmd.AddCommand(&rolesCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "config-path",
		Short: "Print the absolute path to the configuration file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := findConfigPath()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	})
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	rootCmd.PersistentFlags().MarkHidden(FlagOIDCDomain)
	rootCmd.PersistentFlags().MarkHidden(FlagClientID)
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
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %s", err)
		}

		// We don't care about this being cancelled.
		timeout, _ := cmd.Flags().GetInt(FlagTimeout)
		nextCtx, _ := context.WithTimeout(cmd.Context(), time.Duration(timeout)*time.Second)
		cmd.SetContext(ConfigContext(nextCtx, &config))
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
		config := ConfigFromCommand(cmd)
		if err := saveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %s", err)
		}
		return nil
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
