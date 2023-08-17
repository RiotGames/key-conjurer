package main

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	accountsCmd.Flags().StringVar(&identityProvider, "identity-provider", defaultIdentityProvider, "The identity provider to use. Refer to `"+appname+" identity-providers` for more info.")
}

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Prints the list of accounts you have access to.",
	Long:  "Prints the list of accounts you have access to.",
	// Example: appname + " accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if HasTokenExpired(config.Tokens) {
			cmd.PrintErrln("Your session has expired. Please run login again.")
			return nil
		}

		accounts := []Account{}

		var entries []Account
		for _, acc := range accounts {
			entries = append(entries, Account{ID: acc.ID, Name: acc.Name, Alias: generateDefaultAlias(acc.Name)})
		}

		config.UpdateAccounts(entries)
		config.DumpAccounts(os.Stdout)
		return nil
	},
}
