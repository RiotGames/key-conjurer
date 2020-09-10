package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/riotgames/key-conjurer/cli/keyconjurer"

	"github.com/spf13/cobra"
)

var unaliasCmd = &cobra.Command{
	Use:     "unalias <accountName/alias>",
	Short:   "Remove alias from account.",
	Long:    "Removes alias from account. The positional arg can refer to the account by name or alias.",
	Args:    cobra.ExactArgs(1),
	Example: "keyconjurer alias FooAccount Bar",
	Run: func(cmd *cobra.Command, args []string) {
		userData := keyconjurer.Login(keyConjurerRcPath, false)
		account := args[0]
		if err := userData.RemoveAlias(account); err != nil {
			fmt.Println(err.Error())
		}

		if err := userData.Save(); err != nil {
			log.Println(err)
			log.Printf("Unable to save user data to %v\n", keyConjurerRcPath)
			os.Exit(1)
		}
	}}
