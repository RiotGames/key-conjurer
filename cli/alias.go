package main

import (
	"github.com/spf13/cobra"
)

var aliasCmd = cobra.Command{
	Use:     "alias <accountName> <alias>",
	Short:   "Give an account a nickname.",
	Long:    "Alias an account to a nickname so you can refer to the account by the nickname.",
	Args:    cobra.ExactArgs(2),
	Example: "  " + appname + " alias FooAccount Bar",
	Run: func(cmd *cobra.Command, args []string) {
		config := ConfigFromContext(cmd.Context())
		config.Alias(args[0], args[1])
	}}
