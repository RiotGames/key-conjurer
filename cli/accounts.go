package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/riotgames/key-conjurer/internal/api"
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
	accountsCmd.Flags().String(FlagServerAddress, ServerAddress, "The address of the account server. This does not usually need to be changed or specified.")
}

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Prints and optionally refreshes the list of accounts you have access to.",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := ConfigFromCommand(cmd)
		stdOut := cmd.OutOrStdout()
		noRefresh, _ := cmd.Flags().GetBool(FlagNoRefresh)
		if noRefresh {
			config.DumpAccounts(stdOut)
			if q, _ := cmd.Flags().GetBool(FlagQuiet); !q {
				cmd.PrintErrf("--%s was specified - these results may be out of date, and you may not have access to accounts in this list.\n", FlagNoRefresh)
			}
			return nil
		}

		serverAddr, _ := cmd.Flags().GetString(FlagServerAddress)
		serverAddrURI, err := url.Parse(serverAddr)
		if err != nil {
			cmd.PrintErrf("--%s had an invalid value: %s\n", FlagServerAddress, err)
			return nil
		}

		if HasTokenExpired(config.Tokens) {
			cmd.PrintErrln("Your session has expired. Please run login again.")
			config.SaveOAuthToken(nil)
			return nil
		}

		tok := oauth2.Token{
			AccessToken:  config.Tokens.AccessToken,
			RefreshToken: config.Tokens.RefreshToken,
			Expiry:       config.Tokens.Expiry,
			TokenType:    config.Tokens.TokenType,
		}

		accounts, err := refreshAccounts(cmd.Context(), serverAddrURI, &tok)
		if err != nil {
			cmd.PrintErrf("Error refreshing accounts: %s\n", err)
			cmd.PrintErrln("If you don't need to refresh your accounts, consider adding the --no-refresh flag")
			return nil
		}

		config.UpdateAccounts(accounts)
		config.DumpAccounts(stdOut)
		return nil
	},
}

func refreshAccounts(ctx context.Context, serverAddr *url.URL, tok *oauth2.Token) ([]Account, error) {
	uri := serverAddr.ResolveReference(&url.URL{Path: "/v2/applications"})
	httpClient := NewHTTPClient()
	req, _ := http.NewRequestWithContext(ctx, "POST", uri.String(), nil)
	tok.SetAuthHeader(req)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to issue request: %s", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body: %s", err)
	}

	var jsonError api.JSONError
	if resp.StatusCode != http.StatusOK && resp.StatusCode != 0 {
		if err := json.Unmarshal(body, &jsonError); err != nil {
			return nil, errors.New(jsonError.Message)

		}
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var apps []api.Application
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
