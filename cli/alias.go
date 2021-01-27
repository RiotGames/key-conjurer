package main

import (
	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:     "alias <accountName> <alias>",
	Short:   "Give an account a nickname.",
	Long:    "Alias an account to a nickname so you can refer to the account by the nickname.",
	Args:    cobra.ExactArgs(2),
	Example: "keyconjurer alias FooAccount Bar",
	RunE: func(cmd *cobra.Command, args []string) error {
		return userData.NewAlias(args[0], args[1])
	},
}
