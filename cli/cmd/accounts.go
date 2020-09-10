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
	Run: func(cmd *cobra.Command, args []string) {
		userData := keyconjurer.Login(keyConjurerRcPath, false)
		//need update path
		userData.ListAccounts()
	}}
