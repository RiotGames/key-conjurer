package cmd

import (
	"keyconjurer-cli/keyconjurer"
	"log"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Get credentials for Key Conjurer",
	Long: `Login using your AD creds.  This stores encrypted credentials
on the local system`,
	Example: "keyconjurer login",
	Run: func(cmd *cobra.Command, args []string) {
		userData := keyconjurer.Login(keyConjurerRcPath, true)
		if err := userData.Save(); err != nil {
			log.Println(err)
		}
	},
}
