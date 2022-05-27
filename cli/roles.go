package main

import (
	"fmt"

	"github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/spf13/cobra"
)

func init() {
	rolesCmd.Flags().StringVar(&identityProvider, "identity-provider", defaultIdentityProvider, "The identity provider to retrieve roles from")
}

var rolesCmd = cobra.Command{
	Use:   "roles",
	Short: "List all the roles that you can assume when using `" + appname + " get`.",
	RunE: func(*cobra.Command, []string) error {
		switch identityProvider {
		case keyconjurer.AuthenticationProviderOneLogin:
			return fmt.Errorf("roles are not used in the OneLogin authentication provider")
		case keyconjurer.AuthenticationProviderOkta:
			return fmt.Errorf(`You cannot retrieve roles for %q from the command line at this time. Instead, please check the instructions you have received from the team that manages KeyConjurer within your organization`, identityProvider)
		default:
			return fmt.Errorf("unsupported identity provider %q", identityProvider)
		}
	},
}
