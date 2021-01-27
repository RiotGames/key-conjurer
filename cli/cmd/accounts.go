package cmd

import (
	"context"
	"os"

	api "github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/riotgames/key-conjurer/cli/keyconjurer"
	"github.com/spf13/cobra"
)

func init() {
	accountsCmd.Flags().StringVar(&authProvider, "auth-provider", api.AuthenticationProviderOkta, "The authentication provider to use when interacting with the server.")
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

		creds, err := userData.GetCredentials()
		if err != nil {
			return err
		}

		accounts, err := client.ListAccounts(ctx, &keyconjurer.ListAccountsOptions{
			Credentials:            creds,
			AuthenticationProvider: authProvider,
		})

		if err != nil {
			return err
		}

		userData.mergeAccounts(accounts)
		userData.ListAccounts(os.Stdout)
		return nil
	},
}
