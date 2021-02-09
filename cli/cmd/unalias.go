package cmd

import (
	"github.com/riotgames/key-conjurer/cli/keyconjurer"

	"github.com/spf13/cobra"
)

var unaliasCmd = &cobra.Command{
	Use:     "unalias <accountName/alias>",
	Short:   "Remove alias from account.",
	Long:    "Removes alias from account. The positional arg can refer to the account by name or alias.",
	Args:    cobra.ExactArgs(1),
	Example: "keyconjurer alias FooAccount Bar",
	RunE: func(cmd *cobra.Command, args []string) error {
		userData, err := keyconjurer.Login(keyConjurerRcPath, false)
		if err != nil {
			return err
		}

		if err := userData.RemoveAlias(args[0]); err != nil {
			return err
		}

		return userData.Save()
	}}
