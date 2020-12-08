package cmd

import (
	"github.com/riotgames/key-conjurer/cli/keyconjurer"

	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:     "alias <accountName> <alias>",
	Short:   "Give an account a nickname.",
	Long:    "Alias an account to a nickname so you can refer to the account by the nickname.",
	Args:    cobra.ExactArgs(2),
	Example: "keyconjurer alias FooAccount Bar",
	RunE: func(cmd *cobra.Command, args []string) error {
		userData, err := keyconjurer.Login(keyConjurerRcPath, false)
		if err != nil {
			return err
		}

		account := args[0]
		alias := args[1]
		if err := userData.NewAlias(account, alias); err != nil {
			return err
		}

		return userData.Save()
	}}
