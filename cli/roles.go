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
			cmd.PrintErrln("Your session has expired. Please login again.")
			return nil
		}
		client := NewHTTPClient()

		oidcDomain, _ := cmd.Flags().GetString(FlagOIDCDomain)
		clientID, _ := cmd.Flags().GetString(FlagClientID)

		var applicationID = args[0]
		account, ok := config.FindAccount(applicationID)
		if ok {
			applicationID = account.ID
		}

		oauthCfg, err := DiscoverOAuth2Config(cmd.Context(), oidcDomain, clientID)
		if err != nil {
			cmd.PrintErrf("could not discover oauth2  config: %s\n", err)
			return nil
		}

		tok, err := ExchangeAccessTokenForWebSSOToken(cmd.Context(), client, oauthCfg, config.Tokens, applicationID)
		if err != nil {
			cmd.PrintErrf("error exchanging token: %s\n", err)
			return nil
		}

		assertionBytes, err := ExchangeWebSSOTokenForSAMLAssertion(cmd.Context(), client, oidcDomain, tok)
		if err != nil {
			cmd.PrintErrf("failed to fetch SAML assertion: %s\n", err)
			return nil
		}

		samlResponse, err := ParseBase64EncodedSAMLResponse(string(assertionBytes))
		if err != nil {
			cmd.PrintErrf("could not parse assertion: %s\n", err)
			return nil
		}

		for _, name := range ListSAMLRoles(samlResponse) {
			cmd.Println(name)
		}

		return nil
	},
}
