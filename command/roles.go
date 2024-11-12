package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/riotgames/key-conjurer/oauth2"
)

type RolesCommand struct {
	ApplicationID string `arg:""`
}

func (r RolesCommand) Run(globals *Globals, config *Config) error {
	return r.RunContext(context.Background(), globals, config)
}

func (r RolesCommand) RunContext(ctx context.Context, globals *Globals, config *Config) error {
	if HasTokenExpired(config.Tokens) {
		return ErrTokensExpiredOrAbsent
	}

	account, ok := config.FindAccount(r.ApplicationID)
	if ok {
		r.ApplicationID = account.ID
	}

	samlResponse, _, err := oauth2.DiscoverConfigAndExchangeTokenForAssertion(
		ctx,
		config.Tokens.AccessToken,
		config.Tokens.IDToken,
		globals.OIDCDomain,
		globals.ClientID,
		r.ApplicationID,
	)

	if err != nil {
		return err
	}

	for _, name := range listRoles(samlResponse) {
		fmt.Println(name)
	}

	return nil
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
