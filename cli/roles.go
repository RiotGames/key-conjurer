package main

import (
	"time"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/riotgames/key-conjurer/internal"
	"github.com/spf13/cobra"
)

var rolesCmd = cobra.Command{
	Use:   "roles <accountName/alias>",
	Short: "Returns the roles that you have access to in the given account.",
	// Example: appname + " roles <accountName/alias>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if HasTokenExpired(config.Tokens) {
			cmd.PrintErrln("Your session has expired. Please login again.")
			return nil
		}
		client := NewHTTPClient()

		var applicationID = args[0]
		account, ok := config.FindAccount(applicationID)
		if ok {
			applicationID = account.ID
		}

		if account.MostRecentRole != "" && roleName == "" {
			roleName = account.MostRecentRole
		}

		if config.TimeRemaining != 0 && timeRemaining == DefaultTimeRemaining {
			timeRemaining = config.TimeRemaining
		}

		if roleName == "" {
			cmd.PrintErrln("You must specify the --role flag with this command")
			return nil
		}

		var credentials CloudCredentials
		credentials.LoadFromEnv(cloudFlag)
		if credentials.ValidUntil(*account, cloudFlag, time.Duration(timeRemaining)*time.Minute) {
			return echoCredentials(args[0], args[0], credentials, outputType, cloudFlag)
		}

		oauthCfg, _, err := DiscoverOAuth2Config(cmd.Context(), client, OktaDomain)
		if err != nil {
			cmd.PrintErrf("could not discover oauth2  config: %s\n", err)
			return nil
		}

		tok, err := ExchangeAccessTokenForWebSSOToken(cmd.Context(), client, oauthCfg, config.Tokens, applicationID)
		if err != nil {
			cmd.PrintErrf("error exchanging token: %s\n", err)
			return nil
		}

		assertionBytes, err := ExchangeWebSSOTokenForSAMLAssertion(cmd.Context(), client, OktaDomain, tok)
		if err != nil {
			cmd.PrintErrf("failed to fetch SAML assertion: %s\n", err)
			return nil
		}

		assertionStr := string(assertionBytes)

		samlResponse, err := saml.ParseEncodedResponse(assertionStr)
		if err != nil {
			cmd.PrintErrf("could not parse assertion: %s\n", err)
			return nil
		}

		for _, name := range internal.ListRoles(samlResponse) {
			cmd.Println(name)
		}

		return nil
	},
}
