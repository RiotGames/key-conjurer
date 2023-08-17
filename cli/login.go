package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
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
		tok, ok := config.GetOAuthToken()
		if ok && tok.Expiry.After(time.Now()) {
			return nil
		}

		token, err := Login(cmd.Context(), OktaDomain)
		if err != nil {
			return err
		}

		return config.SaveOAuthToken(token)
	},
}

func Login(ctx context.Context, domain string) (*oauth2.Token, error) {
	provider, err := oidc.NewProvider(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("couldn't discover OIDC configuration for %s: %w", OktaDomain, err)
	}

	listener := NewOAuth2Listener()
	go listener.Listen(ctx)

	oauthCfg := oauth2.Config{
		ClientID: ClientID,
		Endpoint: provider.Endpoint(),
		Scopes:   []string{"openid", "email", "okta.apps.read"},
		// TODO: Only use a redirect URL to localhost if the user has supplied the `--open-browser` flag.
		// If they don't, the default behavior should be to display a QR code that the user should scan and follow the device authorization flow instead.
		RedirectURL: listener.Addr,
	}

	state, err := GenerateState()
	if err != nil {
		return nil, err
	}

	codeVerifier, codeChallenge, err := GenerateCodeVerifierAndChallenge()
	if err != nil {
		return nil, err
	}

	url := oauthCfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
	)

	fmt.Printf("Visit the following link in your terminal: %s\n", url)

	// TODO: Allow cancellation with CTRL+C
	// Cobra might already do this for us..
	code, err := listener.WaitForAuthorizationCode(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization code: %w", err)
	}

	token, err := oauthCfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}
