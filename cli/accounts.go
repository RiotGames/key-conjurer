package main

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Prints the list of accounts you have access to.",
	Long:  "Prints the list of accounts you have access to.",
	// Example: appname + " accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if HasTokenExpired(config.Tokens) {
			cmd.PrintErrln("Your session has expired. Please run login again.")
			return nil
		}

		tok := oauth2.Token{
			AccessToken:  config.Tokens.AccessToken,
			RefreshToken: config.Tokens.RefreshToken,
			Expiry:       config.Tokens.Expiry,
			TokenType:    config.Tokens.TokenType,
		}

		httpClient := NewOAuth2Client(cmd.Context(), oauth2.StaticTokenSource(&tok))
		_, client, err := okta.NewClient(
			cmd.Context(),
			okta.WithOrgUrl(oidcDomain),
			okta.WithHttpClient(*httpClient),
			// This is not used - the http client overwrites the tokens when a request is made.
			// It must be specified to satisfy the Okta SDK.
			okta.WithToken("dummy text"),
		)
		if err != nil {
			return err
		}

		accounts, err := FetchAccounts(cmd.Context(), client)
		if errors.Is(err, ErrSessionExpired) {
			config.SaveOAuthToken(nil)
			cmd.PrintErrln("Your session has expired. Please run login again.")
			return nil
		} else if err != nil {
			return err
		}

		config.UpdateAccounts(accounts)
		config.DumpAccounts(os.Stdout)
		return nil
	},
}

var ErrSessionExpired = errors.New("session expired")

func FetchAccounts(ctx context.Context, client *okta.Client) ([]Account, error) {
	apps, resp, err := client.Application.ListApplications(ctx, query.NewQueryParams())
	if err != nil {
		if resp.StatusCode == http.StatusUnauthorized {
			// Tokens expired.
			return nil, ErrSessionExpired
		}
		return nil, err
	}

	var entries []Account
	for _, app := range apps {
		app, ok := app.(*okta.Application)
		if !ok {
			continue
		}

		entries = append(entries, Account{ID: app.Id, Name: app.Label, Alias: generateDefaultAlias(app.Label)})
	}
	return entries, nil
}
