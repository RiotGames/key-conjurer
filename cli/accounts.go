package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/pkg/httputil"
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
		quiet, _ := cmd.Flags().GetBool("quiet")
		noRefresh, _ := cmd.Flags().GetBool(FlagNoRefresh)
		if !noRefresh {
			serverAddr, _ := cmd.Flags().GetString(FlagServerAddress)
			serverAddrUri, err := url.Parse(serverAddr)
			if err != nil {
				cmd.PrintErrf("--%s had an invalid value: %s", FlagServerAddress, err)
				return nil
			}

			accounts, err := refreshAccounts(cmd.Context(), serverAddrUri, config.Tokens)
			if errors.Is(err, ErrSessionExpired) {
				cmd.PrintErrln("Your session has expired. Please run login again.")
				config.SaveOAuthToken(nil)
				return nil
			} else if err != nil {
				cmd.PrintErrf("Error refreshing accounts: %s", err)
				cmd.PrintErrln("If you don't need to refresh your accounts, consider adding the --no-refresh flag")
				return nil
			}

			config.UpdateAccounts(accounts)
		}

		config.DumpAccounts(os.Stdout)
		if noRefresh && !quiet {
			cmd.PrintErrf("--%s was specified - these results may be out of date, and you may not have access to accounts in this list.", FlagNoRefresh)
		}

		return nil
	},
}

func refreshAccounts(ctx context.Context, serverAddr *url.URL, tokens *TokenSet) ([]Account, error) {
	if HasTokenExpired(tokens) {
		return nil, ErrSessionExpired
	}

	tok := oauth2.Token{
		AccessToken:  config.Tokens.AccessToken,
		RefreshToken: config.Tokens.RefreshToken,
		Expiry:       config.Tokens.Expiry,
		TokenType:    config.Tokens.TokenType,
	}

	uri := serverAddr.ResolveReference(&url.URL{Path: "/v2/applications"})
	httpClient := NewOAuth2Client(ctx, oauth2.StaticTokenSource(&tok))
	req, _ := http.NewRequestWithContext(ctx, "POST", uri.String(), nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to issue request: %s", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body: %s", err)
	}

	var jsonError httputil.JSONError
	if resp.StatusCode != http.StatusOK {
		if err := json.Unmarshal(body, &jsonError); err != nil {
			return nil, errors.New(jsonError.Message)

		}
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var apps []core.Application
	if err := json.Unmarshal(body, &apps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal applications: %w", err)
	}

	entries := make([]Account, len(apps))
	for idx, app := range apps {
		entries[idx] = Account{
			ID:    app.ID,
			Name:  app.Name,
			Alias: generateDefaultAlias(app.Name),
		}
	}

	return entries, nil
}
