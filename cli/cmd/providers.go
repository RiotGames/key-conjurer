package cmd

import (
	"context"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/riotgames/key-conjurer/cli/keyconjurer"
	"github.com/spf13/cobra"
)

var providersCmd = &cobra.Command{
	Use:     "providers",
	Short:   "List authentication providers you may use.",
	Example: "keyconjurer providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := newClient()
		if err != nil {
			return err
		}

		providers, err := client.ListProviders(ctx, &keyconjurer.ListProvidersOptions{})
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
