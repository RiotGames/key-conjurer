package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slog"
)

var (
	FlagURLOnly   = "url-only"
	FlagNoBrowser = "no-browser"
)

func init() {
	loginCmd.Flags().BoolP(FlagURLOnly, "u", false, "Print only the URL to visit rather than a user-friendly message")
	loginCmd.Flags().BoolP(FlagNoBrowser, "b", false, "Do not open a browser window, printing the URL instead")
}

// ShouldUseMachineOutput indicates whether or not we should write to standard output as if the user is a machine.
//
// What this means is implementation specific, but this usually indicates the user is trying to use this program in a script and we should avoid user-friendly output messages associated with values a user might find useful.
func ShouldUseMachineOutput(flags *pflag.FlagSet) bool {
	quiet, _ := flags.GetBool(FlagQuiet)
	fi, _ := os.Stdout.Stat()
	isPiped := fi.Mode()&os.ModeCharDevice == 0
	return isPiped || quiet
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with KeyConjurer.",
	Long:  "Login to KeyConjurer using OAuth2. You will be required to open the URL printed to the console or scan a QR code.",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := ConfigFromCommand(cmd)
		if !HasTokenExpired(config.Tokens) {
			return nil
		}

		oidcDomain, _ := cmd.Flags().GetString(FlagOIDCDomain)
		clientID, _ := cmd.Flags().GetString(FlagClientID)
		urlOnly, _ := cmd.Flags().GetBool(FlagURLOnly)
		noBrowser, _ := cmd.Flags().GetBool(FlagNoBrowser)
		command := LoginCommand{
			Config:        config,
			OIDCDomain:    oidcDomain,
			ClientID:      clientID,
			MachineOutput: ShouldUseMachineOutput(cmd.Flags()) || urlOnly,
			NoBrowser:     noBrowser,
		}

		return command.Execute(cmd.Context())
	},
}

type LoginCommand struct {
	Config        *Config
	OIDCDomain    string
	ClientID      string
	MachineOutput bool
	NoBrowser     bool
}

func (c LoginCommand) Execute(ctx context.Context) error {
	oauthCfg, err := DiscoverOAuth2Config(ctx, c.OIDCDomain, c.ClientID)
	if err != nil {
		return err
	}

	handler := RedirectionFlowHandler{
		Config:       oauthCfg,
		OnDisplayURL: openBrowserToURL,
	}

	if c.NoBrowser {
		if c.MachineOutput {
			handler.OnDisplayURL = printURLToConsole
		} else {
			handler.OnDisplayURL = friendlyPrintURLToConsole
		}
	}

	state := GenerateState()
	challenge := GeneratePkceChallenge()
	token, err := handler.HandlePendingSession(ctx, challenge, state)
	if err != nil {
		return err
	}

	return c.Config.SaveOAuthToken(token)
}

func friendlyPrintURLToConsole(url string) error {
	fmt.Printf("Visit the following link in your terminal: %s\n", url)
	return nil
}

func openBrowserToURL(url string) error {
	slog.Debug("trying to open browser window", slog.String("url", url))
	return browser.OpenURL(url)
}
