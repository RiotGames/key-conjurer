package cmd

import (
	"github.com/riotgames/key-conjurer/cli/keyconjurer"

	"github.com/spf13/cobra"
)

var (
	updateAccounts bool
)

func init() {
	accountsCmd.Flags().BoolVar(&updateAccounts, "update", false, "Used to update accounts")
}

var accountsCmd = &cobra.Command{
	Use:     "accounts",
	Short:   "Prints the list of accounts you have access to.",
	Long:    "Prints the list of accounts you have access to.",
	Example: "keyconjurer accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		userData, err := keyconjurer.Login(keyConjurerRcPath, false)
		if err != nil {
			return err
		}

		//need update path
		return userData.ListAccounts()
	}}
