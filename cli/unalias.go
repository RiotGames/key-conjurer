package main

import (
	"github.com/spf13/cobra"
)

var unaliasCmd = cobra.Command{
	Use:     "unalias <accountName/alias>",
	Short:   "Remove alias from account.",
	Args:    cobra.ExactArgs(1),
	Example: "  " + appname + " unalias bar",
	Run: func(cmd *cobra.Command, args []string) {
		config.Unalias(args[0])
	}}
