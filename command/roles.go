package command

import (
	"strings"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/riotgames/key-conjurer/pkg/oauth2cli"
	"github.com/spf13/cobra"
)

var rolesCmd = cobra.Command{
	Use:   "roles <accountName/alias>",
	Short: "Returns the roles that you have access to in the given account.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config := ConfigFromCommand(cmd)
		oidcDomain, _ := cmd.Flags().GetString(FlagOIDCDomain)
		clientID, _ := cmd.Flags().GetString(FlagClientID)

		var applicationID = args[0]
		account, ok := config.FindAccount(applicationID)
		if ok {
			applicationID = account.ID
		}

		samlResponse, _, err := oauth2cli.DiscoverConfigAndExchangeTokenForAssertion(cmd.Context(), &keychainTokenSource{}, oidcDomain, clientID, applicationID)
		if err != nil {
			return err
		}

		for _, name := range listRoles(samlResponse) {
			cmd.Println(name)
		}

		return nil
	},
}

type roleProviderPair struct {
	RoleARN     string
	ProviderARN string
}

func getARN(value string) roleProviderPair {
	var p roleProviderPair
	roles := strings.Split(value, ",")
	if len(roles) >= 2 {
		if strings.Contains(roles[0], "saml-provider/") {
			p.ProviderARN = roles[0]
			p.RoleARN = roles[1]
		} else {
			p.ProviderARN = roles[1]
			p.RoleARN = roles[0]
		}
	}
	return p
}

func findRoleInSAML(roleName string, response *saml.Response) (roleProviderPair, bool) {
	if response == nil {
		return roleProviderPair{}, false
	}

	roleURL := "https://aws.amazon.com/SAML/Attributes/Role"
	roleSubstr := "role/"
	attrs := response.GetAttributeValues(roleURL)

	if len(attrs) == 0 {
		return roleProviderPair{}, false
	}

	var pairs []roleProviderPair
	for _, v := range response.GetAttributeValues(roleURL) {
		pairs = append(pairs, getARN(v))
	}

	if len(pairs) == 0 {
		return roleProviderPair{}, false
	}

	var pair roleProviderPair
	for _, p := range pairs {
		idx := strings.Index(p.RoleARN, roleSubstr)
		parts := strings.Split(p.RoleARN[idx:], "/")
		if strings.EqualFold(parts[1], roleName) {
			pair = p
		}
	}

	if pair.RoleARN == "" {
		return roleProviderPair{}, false
	}

	return pair, true
}

func listRoles(response *saml.Response) []string {
	if response == nil {
		return nil
	}

	roleURL := "https://aws.amazon.com/SAML/Attributes/Role"
	roleSubstr := "role/"

	var names []string
	for _, v := range response.GetAttributeValues(roleURL) {
		p := getARN(v)
		idx := strings.Index(p.RoleARN, roleSubstr)
		parts := strings.Split(p.RoleARN[idx:], "/")
		names = append(names, parts[1])
	}

	return names
}
