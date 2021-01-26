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
	rolesCmd.Flags().StringVar(&authProvider, "auth-provider", api.AuthenticationProviderOkta, "The authentication provider to use when interacting with the server.")
}

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "List roles in Key Conjurer for the given account.",
	Long: `List roles in KeyConjurer for the given account.

You must be logged in.`,
	Example: "keyconjurer roles",
	Args:    cobra.MaximumNArgs(1),
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

		roles, err := client.ListRoles(ctx, &keyconjurer.ListRolesOptions{
			AuthenticationProvider: authProvider,
			Credentials:            creds,
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
