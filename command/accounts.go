package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/riotgames/key-conjurer/internal/api"
	"golang.org/x/oauth2"
)

var (
	FlagNoRefresh     = "no-refresh"
	FlagServerAddress = "server-address"
	ErrSessionExpired = errors.New("session expired")
)

type AccountsCommand struct {
	Refresh       bool   `help:"Refresh the list of accounts." default:"true" negatable:""`
	ServerAddress string `help:"The address of the account server. This does not usually need to be changed or specified." hidden:"" env:"KEYCONJURER_SERVER_ADDRESS" default:"${server_address}"`
}

func (a AccountsCommand) Help() string {
	return "Prints and optionally refreshes the list of accounts you have access to."
}

func (a AccountsCommand) RunContext(ctx context.Context, globals *Globals, config *Config) error {
	loud := isPiped() || globals.Quiet
	if !a.Refresh {
		config.DumpAccounts(os.Stdout, loud)

		if loud {
			// intentionally uses Fprintf was a warning
			fmt.Fprintf(os.Stderr, "--no-refresh was specified - these results may be out of date, and you may not have access to accounts in this list.\n")
		}

		return nil
	}

	serverAddrURI, err := url.Parse(a.ServerAddress)
	if err != nil {
		return genericError{
			ExitCode: ExitCodeValueError,
			Message:  fmt.Sprintf("--%s had an invalid value: %s\n", FlagServerAddress, err),
		}
	}

	if HasTokenExpired(config.Tokens) {
		return ErrTokensExpiredOrAbsent
	}

	accounts, err := refreshAccounts(ctx, serverAddrURI, config.Tokens)
	if err != nil {
		return fmt.Errorf("error refreshing accounts: %w", err)
	}

	config.UpdateAccounts(accounts)
	config.DumpAccounts(os.Stdout, loud)
	return nil
}

func (a AccountsCommand) Run(globals *Globals, config *Config) error {
	return a.RunContext(context.Background(), globals, config)
}

func refreshAccounts(ctx context.Context, serverAddr *url.URL, ts oauth2.TokenSource) ([]Account, error) {
	client := oauth2.NewClient(ctx, ts)
	uri := serverAddr.ResolveReference(&url.URL{Path: "/v2/applications"})
	req, _ := http.NewRequestWithContext(ctx, "POST", uri.String(), nil)
	resp, err := client.Do(req)
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
