package command

import (
	"testing"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/stretchr/testify/require"
)

func Test_findRoleInSAML_DoesntBreakIfYouHaveMultipleRoles(t *testing.T) {
	var resp saml.Response
	resp.AddAttribute("https://aws.amazon.com/SAML/Attributes/Role", "arn:cloud:iam::1234:saml-provider/Okta,arn:cloud:iam::1234:role/Admin")
	resp.AddAttribute("https://aws.amazon.com/SAML/Attributes/Role", "arn:cloud:iam::1234:saml-provider/Okta,arn:cloud:iam::1234:role/Power")
	pair, err := findRoleInSAML("Power", &resp)
	require.True(t, err)
	require.Equal(t, "arn:cloud:iam::1234:saml-provider/Okta", pair.ProviderARN)
	require.Equal(t, "arn:cloud:iam::1234:role/Power", pair.RoleARN)
	pair, err = findRoleInSAML("Admin", &resp)
	require.True(t, err)
	require.Equal(t, "arn:cloud:iam::1234:saml-provider/Okta", pair.ProviderARN)
	require.Equal(t, "arn:cloud:iam::1234:role/Admin", pair.RoleARN)
}
