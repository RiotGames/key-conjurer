package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/riotgames/key-conjurer/pkg/oauth2device"
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

		token, err := Login(cmd.Context(), OktaDomain, true)
		if err != nil {
			return err
		}

		return config.SaveOAuthToken(token)
	},
}

func Login(ctx context.Context, domain string, useDeviceFlow bool) (*oauth2.Token, error) {
	provider, err := oidc.DiscoverProvider(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("couldn't discover OIDC configuration for %s: %w", OktaDomain, err)
	}

	state, err := GenerateState()
	if err != nil {
		return nil, err
	}

	codeVerifier, codeChallenge, err := GenerateCodeVerifierAndChallenge()
	if err != nil {
		return nil, err
	}

	oauthCfg := oauth2.Config{
		ClientID: ClientID,
		Endpoint: provider.Endpoint(),
		Scopes:   []string{"openid", "profile", "offline_access", "okta.apps.read"},
	}

	// The device flow and the redirect flow are almost indistinguishable from a user point of view.
	//
	// The device flow should be preferred as it gives the user the option to open a browser on their mobile device or their terminal, whereas the redirect flow requires opening a browser on the current machine.
	if useDeviceFlow && oidc.SupportsDeviceFlow(provider) {
		oauthDeviceCfg := oauth2device.Config{
			Config:         &oauthCfg,
			DeviceEndpoint: provider.DeviceAuthorizationEndpoint(),
		}

		code, err := oauth2device.RequestDeviceCode(http.DefaultClient, &oauthDeviceCfg)
		if err != nil {
			return nil, err
		}
		// TODO: Display a QR code that automatically does this for the user.
		fmt.Printf("Visit %s\n", code.VerificationURLComplete)
		return oauth2device.WaitForDeviceAuthorization(http.DefaultClient, &oauthDeviceCfg, code)
	} else {
		listener := NewOAuth2Listener()
		go listener.Listen(ctx)
		oauthCfg.RedirectURL = "http://localhost:8080"
		url := oauthCfg.AuthCodeURL(state,
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		)

		fmt.Printf("Visit the following link in your terminal: %s\n", url)
		code, err := listener.WaitForAuthorizationCode(ctx, state)
		if err != nil {
			return nil, fmt.Errorf("failed to get authorization code: %w", err)
		}

		return oauthCfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}
}
