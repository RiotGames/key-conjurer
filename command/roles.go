package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/riotgames/key-conjurer/pkg/oauth2cli"
	"github.com/urfave/cli/v3"
)

var rolesCmd = &cli.Command{
	Name:  "roles",
	Usage: "Returns the roles that you have access to in the given account.",
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name: "accountName/alias",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		config := ConfigFromContext(ctx)
		oidcDomain := cmd.String(FlagOIDCDomain)
		clientID := cmd.String(FlagClientID)
		// See positionalArgs' doc comment in cliargs.go: cmd.Args() is empty
		// here because the Arguments field above already consumed the
		// "accountName/alias" value; cmd.Args().First() always returns "".
		var applicationID = cmd.StringArg("accountName/alias")
		if applicationID == "" {
			return cli.Exit("accountName/alias is required", 1)
		}

		account, ok := config.FindAccount(applicationID)
		if ok {
			applicationID = account.ID
		}

		samlResponse, _, err := oauth2cli.DiscoverConfigAndExchangeTokenForAssertion(ctx, &keychainTokenSource{}, oidcDomain, clientID, applicationID)
		if err != nil {
			return err
		}

		for _, name := range listRoles(samlResponse) {
			fmt.Fprintln(cmd.Writer, name)
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
