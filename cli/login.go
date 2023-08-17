package main

import (
	"context"
	"os"

	"github.com/riotgames/key-conjurer/pkg/oidc"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	ClientID   = os.Getenv("OKTA_CLIENT_ID")
	OktaDomain = os.Getenv("OKTA_DOMAIN")
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with KeyConjurer.",
	Long:  "Login to KeyConjurer using OAuth2. You will be required to open the URL printed to the console or scan a QR code.",
	// Example: appname + " login",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !HasTokenExpired(config.Tokens) {
			return nil
		}

		token, err := Login(cmd.Context(), OktaDomain, false)
		if err != nil {
			return err
		}

		return config.SaveOAuthToken(token)
	},
}

func Login(ctx context.Context, domain string, useDeviceFlow bool) (*oauth2.Token, error) {
	oauthCfg, provider, err := DiscoverOAuth2Config(ctx, domain)
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

	// The device flow and the redirect flow are almost indistinguishable from a user point of view.
	//
	// The device flow should be preferred as it gives the user the option to open a browser on their mobile device or their terminal, whereas the redirect flow requires opening a browser on the current machine.
	if useDeviceFlow && oidc.SupportsDeviceFlow(provider) {
		// TODO: This flow is not currently exposing ID tokens so it cannot be used to access an SSO account
		return DeviceAuthorizationFlow(provider, oauthCfg)
	} else {
		return RedirectionFlow(ctx, oauthCfg, state, codeChallenge, codeVerifier)
	}
}
