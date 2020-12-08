package cmd

import (
	"github.com/riotgames/key-conjurer/cli/keyconjurer"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Get credentials for Key Conjurer",
	Long: `Login using your AD creds.  This stores encrypted credentials
on the local system`,
	Example: "keyconjurer login",
	RunE: func(cmd *cobra.Command, args []string) error {
		userData, err := keyconjurer.Login(keyConjurerRcPath, false)
		if err != nil {
			return err
		}

		return userData.Save()
	},
}
