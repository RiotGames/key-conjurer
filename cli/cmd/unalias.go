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
		var ud keyconjurer.UserData
		if err := ud.LoadFromFile(keyConjurerRcPath); err != nil {
			return err
		}

		if !ud.RemoveAlias(args[0]) {
			// No need to save if no alias was removed
			return nil
		}

		return ud.Save()
	},
}
