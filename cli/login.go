package main

import (
	"context"
	"net/http"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
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

		oidcDomain, _ := cmd.Flags().GetString(FlagOIDCDomain)
		clientID, _ := cmd.Flags().GetString(FlagClientID)
		token, err := Login(cmd.Context(), NewHTTPClient(), oidcDomain, clientID)
		if err != nil {
			return err
		}

		return config.SaveOAuthToken(token)
	},
}

func Login(ctx context.Context, client *http.Client, domain, clientID string) (*oauth2.Token, error) {
	oauthCfg, _, err := DiscoverOAuth2Config(ctx, client, domain, clientID)
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

	return RedirectionFlow(ctx, oauthCfg, state, codeChallenge, codeVerifier)
}
