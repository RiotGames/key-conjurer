package cmd

import (
	"context"
	"os"

	"github.com/olekukonko/tablewriter"
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
		// TODO: List aliases
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

		tw := tablewriter.NewWriter(os.Stdout)
		tw.SetHeader([]string{"Account ID", "Account Name"})
		for _, account := range accounts {
			tw.Append([]string{account.ID, account.Name})
		}

		tw.Render()
		return nil
	},
}
