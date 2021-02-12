package main

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var providersCmd = cobra.Command{
	Use:   "identity-providers",
	Short: "List identity providers you may use.",
	Long: fmt.Sprintf(`List all identity providers that KeyConjurer supports through which the user may authenticate.

If KeyConjurer supports multiple providers, you may specify one you wish to use with the --identity-provider flag.

If you do not specify an --identity-provider flag for the commands that support it (get, login, accounts) a default identity provider will be chosen for you (default: %q).
`, defaultIdentityProvider),
	Example: "keyconjurer identity-providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := newClient()
		if err != nil {
			return err
		}

		providers, err := client.ListProviders(ctx, &ListProvidersOptions{})
		if err != nil {
			return err
		}

		tw := tablewriter.NewWriter(os.Stdout)
		tw.SetHeader([]string{"ID"})
		for _, provider := range providers {
			tw.Append([]string{provider.ID})
		}

		tw.Render()
		return nil
	},
}
