package main

import (
	"github.com/spf13/cobra"
)

var unaliasCmd = &cobra.Command{
	Use:     "unalias <accountName/alias>",
	Short:   "Remove alias from account.",
	Long:    "Removes alias from account. The positional arg can refer to the account by name or alias.",
	Args:    cobra.ExactArgs(1),
	Example: "keyconjurer alias FooAccount Bar",
	Run: func(cmd *cobra.Command, args []string) {
		config.RemoveAlias(args[0])
	},
}
