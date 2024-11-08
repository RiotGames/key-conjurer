package main

import (
	"github.com/spf13/cobra"
)

var rolesCmd = cobra.Command{
	Use:   "roles <accountName/alias>",
	Short: "Returns the roles that you have access to in the given account.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config := ConfigFromCommand(cmd)
		if HasTokenExpired(config.Tokens) {
			return ErrTokensExpiredOrAbsent
		}

		oidcDomain, _ := cmd.Flags().GetString(FlagOIDCDomain)
		clientID, _ := cmd.Flags().GetString(FlagClientID)

		var applicationID = args[0]
		account, ok := config.FindAccount(applicationID)
		if ok {
			applicationID = account.ID
		}

		samlResponse, _, err := DiscoverConfigAndExchangeTokenForAssertion(cmd.Context(), config.Tokens, oidcDomain, clientID, applicationID)
		if err != nil {
			return err
		}

		for _, name := range ListSAMLRoles(samlResponse) {
			cmd.Println(name)
		}

		return nil
	},
}
