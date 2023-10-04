package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
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

		var outputMode LoginOutputMode = LoginOutputModeBrowser{}
		if noBrowser, _ := cmd.Flags().GetBool(FlagNoBrowser); noBrowser {
			if ShouldUseMachineOutput(cmd.Flags()) || urlOnly {
				outputMode = LoginOutputModeURLOnly{}
			} else {
				outputMode = LoginOutputModeHumanFriendlyMessage{}
			}
		}

		token, err := Login(cmd.Context(), oidcDomain, clientID, outputMode)
		if err != nil {
			return err
		}

		return config.SaveOAuthToken(token)
	},
}

func Login(ctx context.Context, domain, clientID string, outputMode LoginOutputMode) (*oauth2.Token, error) {
	oauthCfg, err := DiscoverOAuth2Config(ctx, domain, clientID)
	if err != nil {
		return nil, err
	}

	state, err := GenerateState()
	if err != nil {
		return nil, err
	}

	codeVerifier, codeChallenge, err := GenerateCodeVerifierAndChallenge()
	if err != nil {
		return nil, err
	}

	return RedirectionFlow(ctx, oauthCfg, state, codeChallenge, codeVerifier, outputMode)
}

type LoginOutputMode interface {
	PrintURL(url string) error
}

type LoginOutputModeBrowser struct{}

func (LoginOutputModeBrowser) PrintURL(url string) error {
	slog.Debug("trying to open browser window", slog.String("url", url))
	return browser.OpenURL(url)
}

type LoginOutputModeURLOnly struct{}

func (LoginOutputModeURLOnly) PrintURL(url string) error {
	fmt.Fprintln(os.Stdout, url)
	return nil
}

type LoginOutputModeHumanFriendlyMessage struct{}

func (LoginOutputModeHumanFriendlyMessage) PrintURL(url string) error {
	fmt.Printf("Visit the following link in your terminal: %s\n", url)
	return nil
}
