package main

import (
	"context"
	"os"

	"github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/spf13/cobra"
)

func init() {
	accountsCmd.Flags().StringVar(&authProvider, "provider", keyconjurer.AuthenticationProviderOkta, "The authentication provider to use when interacting with the server.")
}

var accountsCmd = &cobra.Command{
	Use:     "accounts",
	Short:   "Prints the list of accounts you have access to.",
	Long:    "Prints the list of accounts you have access to.",
	Example: "keyconjurer accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := newClient()
		if err != nil {
			return err
		}

		creds, err := config.GetCredentials()
		if err != nil {
			return err
		}

		accounts, err := client.ListAccounts(ctx, &ListAccountsOptions{
			Credentials:            creds,
			AuthenticationProvider: authProvider,
		})

		if err != nil {
			return err
		}

		var entries []Account
		for _, acc := range accounts {
			entries = append(entries, Account{ID: acc.ID, Name: acc.Name, Alias: generateDefaultAlias(acc.Name)})
		}

		cfg := config.Accounts
		cfg.ReplaceWith(entries)
		cfg.WriteTable(os.Stdout)
		return nil
	},
}
