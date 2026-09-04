package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/riotgames/key-conjurer/internal/api"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2"
)

var (
	FlagNoRefresh     = "no-refresh"
	FlagServerAddress = "server-address"

	ErrSessionExpired = errors.New("session expired")
)

var accountsCmd = &cli.Command{
	Name:  "accounts",
	Usage: "Fetches and displays the list of accounts you have access to.",
	Flags: []cli.Flag{
		&cli.BoolWithInverseFlag{
			Name:  FlagNoRefresh,
			Usage: "Indicate that the account list should not be refreshed when executing this command. This is useful if you're not able to reach the account server.",
		},
		&cli.StringFlag{
			Name:  FlagServerAddress,
			Value: ServerAddress,
			Usage: "The address of the account server. This does not usually need to be changed or specified.",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		config := ConfigFromContext(ctx)
		noRefresh := cmd.Bool(FlagNoRefresh)
		loud := !ShouldUseMachineOutput(cmd)
		if noRefresh {
			config.DumpAccounts(cmd.Writer, loud)

			if loud {
				fmt.Fprintf(cmd.ErrWriter, "--%s was specified - these results may be out of date, and you may not have access to accounts in this list.\n", FlagNoRefresh)
			}

			return nil
		}

		serverAddr := cmd.String(FlagServerAddress)
		serverAddrURI, err := url.Parse(serverAddr)
		if err != nil {
			return genericError{
				ExitCode: ExitCodeValueError,
				Message:  fmt.Sprintf("--%s had an invalid value: %s\n", FlagServerAddress, err),
			}
		}

		accounts, err := refreshAccounts(ctx, serverAddrURI, &keychainTokenSource{})
		if err != nil {
			return fmt.Errorf("error refreshing accounts: %w", err)
		}

		config.UpdateAccounts(accounts)
		config.DumpAccounts(cmd.Writer, loud)
		return nil
	},
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
