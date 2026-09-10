package command

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/urfave/cli/v3"
)

var (
	FlagOIDCDomain = "oidc-domain"
	FlagClientID   = "client-id"
	FlagQuiet      = "quiet"
	FlagTimeout    = "timeout"
)

func init() {
	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Fprintln(cmd.Writer, cmd.Version)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cli.Command{
	Name:    "keyconjurer",
	Version: fmt.Sprintf("keyconjurer-%s-%s %s (%s)", runtime.GOOS, runtime.GOARCH, Version, BuildTimestamp),
	Usage:   "Retrieve temporary cloud credentials.",
	Description: `KeyConjurer retrieves temporary credentials from Okta with the assistance of an optional API.

To get started run the following commands:
  keyconjurer login
  keyconjurer accounts
  keyconjurer get <accountName>
`,

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:   FlagOIDCDomain,
			Usage:  "The OIDC domain to use for authentication",
			Hidden: true,
			Value:  OIDCDomain,
		},
		&cli.StringFlag{
			Name:   FlagClientID,
			Usage:  "The client ID to use for authentication",
			Hidden: true,
			Value:  ClientID,
		},
		&cli.DurationFlag{
			Name:  FlagTimeout,
			Usage: "The amount of time to wait for keyconjurer to respond",
			Value: 2 * time.Minute,
		},
		&cli.BoolFlag{
			Name:  FlagQuiet,
			Usage: "Tells the CLI to be quiet; stdout will not contain human-readable informational messages",
		},
	},

	Commands: []*cli.Command{
		loginCmd,
		getCmd,
		accountsCmd,
		rolesCmd,
		cfgCmd,
		switchCmd,
		aliasCmd,
		unaliasCmd,
	},

	Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		config, err := loadConfig()
		if err != nil {
			return ctx, fmt.Errorf("failed to load config: %s", err)
		}

		// We don't care about this being cancelled.
		timeout := cmd.Duration(FlagTimeout)
		nextCtx, _ := context.WithTimeout(ctx, timeout)
		return context.WithValue(nextCtx, ctxKeyConfig{}, &config), nil
	},

	After: func(ctx context.Context, _ *cli.Command) error {
		config := ctx.Value(ctxKeyConfig{}).(*Config)
		if err := saveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %s", err)
		}
		return nil
	},
}

func Execute(ctx context.Context, args []string) error {
	client := &http.Client{Transport: LogRoundTripper{http.DefaultTransport}}
	ctx = oidc.ClientContext(ctx, client)
	return rootCmd.Run(ctx, args)
}
