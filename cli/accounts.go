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

var (
	FlagNoRefresh     = "no-refresh"
	FlagServerAddress = "server-address"

	ErrSessionExpired = errors.New("session expired")
)

func init() {
	accountsCmd.Flags().Bool(FlagNoRefresh, false, "Indicate that the account list should not be refreshed when executing this command. This is useful if you're not able to reach the account server.")
	// TODO: Replace the address
	accountsCmd.Flags().String(FlagServerAddress, "http://localhost:4000", "The address of the account server. This does not usually need to be changed or specified.")
}

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Prints and optionally refreshes the list of accounts you have access to.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if val, _ := cmd.Flags().GetBool(FlagNoRefresh); !val {
			accounts, err := refreshAccounts(cmd.Context(), config.Tokens)
			if errors.Is(err, ErrSessionExpired) {
				cmd.PrintErrln("Your session has expired. Please run login again.")
				config.SaveOAuthToken(nil)
				return nil
			} else if err != nil {
				return err
			}

			config.UpdateAccounts(accounts)
		}

		config.DumpAccounts(os.Stdout)
		return nil
	},
}

func refreshAccounts(ctx context.Context, tokens *TokenSet) ([]Account, error) {
	if HasTokenExpired(tokens) {
		return nil, ErrSessionExpired
	}

	tok := oauth2.Token{
		AccessToken:  config.Tokens.AccessToken,
		RefreshToken: config.Tokens.RefreshToken,
		Expiry:       config.Tokens.Expiry,
		TokenType:    config.Tokens.TokenType,
	}

	httpClient := NewOAuth2Client(ctx, oauth2.StaticTokenSource(&tok))
	_, client, err := okta.NewClient(
		ctx,
		okta.WithOrgUrl(oidcDomain),
		okta.WithHttpClient(*httpClient),
		// This is not used - the http client overwrites the tokens when a request is made.
		// It must be specified to satisfy the Okta SDK.
		okta.WithToken("dummy text"),
	)

	if err != nil {
		return nil, ErrSessionExpired
	}

	return FetchAccounts(ctx, client)
}

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
