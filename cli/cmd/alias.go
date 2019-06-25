package cmd

import (
	"log"

	"keyconjurer-cli/keyconjurer"

	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:     "alias <accountName> <alias>",
	Short:   "Give an account a nickname.",
	Long:    "Alias an account to a nickname so you can refer to the account by the nickname.",
	Args:    cobra.ExactArgs(2),
	Example: "keyconjurer alias FooAccount Bar",
	Run: func(cmd *cobra.Command, args []string) {
		userData := keyconjurer.Login(keyConjurerRcPath, false)
		account := args[0]
		alias := args[1]
		if err := userData.NewAlias(account, alias); err != nil {
			log.Println(err.Error())
		}

		if err := userData.Save(); err != nil {
			log.Println(err)
		}
	}}
