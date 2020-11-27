package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"

	api "github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/riotgames/key-conjurer/cli/keyconjurer"
	"github.com/spf13/cobra"
)

func init() {
	rolesCmd.Flags().StringVar(&authProvider, "auth-provider", api.AuthenticationProviderOkta, "The authentication provider to use when interacting with the server.")
}

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "List roles in Key Conjurer for the given account.",
	Long: `List roles in KeyConjurer for the given account.

You must be logged in.`,
	Example: "keyconjurer roles [account-name]",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := newClient()
		if err != nil {
			return err
		}

		creds, err := loadCredentialsFromFile()
		if err != nil {
			return err
		}

		// TODO: Allow users to use either an account name or an account id
		// ListRoles endpoint only supports an account ID, so we will need to retrieve account names beforehand
		// We could cache them in userdata as is currently the case
		var accountID string
		if len(args) == 1 {
			accountID = strings.TrimSpace(args[0])
		}

		roles, err := client.ListRoles(ctx, &keyconjurer.ListRolesOptions{
			AuthenticationProvider: authProvider,
			Credentials:            creds,
			AccountID:              accountID,
		})

		if err != nil {
			return err
		}

		tw := tablewriter.NewWriter(os.Stdout)
		tw.SetHeader([]string{"Account Name", "Role Name"})
		for _, role := range roles {
			tw.Append([]string{role.AccountName, role.RoleName})
		}

		tw.Render()
		return nil
	},
}
