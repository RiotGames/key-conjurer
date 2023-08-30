package internal

import (
	"testing"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/stretchr/testify/require"
)

func TestAwsFindRoleDoesntBreakIfYouHaveMultipleRoles(t *testing.T) {
	resp := saml.Response{}
	resp.AddAttribute("https://aws.amazon.com/SAML/Attributes/Role", "arn:cloud:iam::1234:saml-provider/Okta,arn:cloud:iam::1234:role/Admin")
	resp.AddAttribute("https://aws.amazon.com/SAML/Attributes/Role", "arn:cloud:iam::1234:saml-provider/Okta,arn:cloud:iam::1234:role/Power")
	pair, _, err := FindRole("Power", &resp)
	require.True(t, err)
	require.Equal(t, "arn:cloud:iam::1234:saml-provider/Okta", pair.ProviderARN)
	require.Equal(t, "arn:cloud:iam::1234:role/Power", pair.RoleARN)
	pair, _, err = FindRole("Admin", &resp)
	require.True(t, err)
	require.Equal(t, "arn:cloud:iam::1234:saml-provider/Okta", pair.ProviderARN)
	require.Equal(t, "arn:cloud:iam::1234:role/Admin", pair.RoleARN)
}

func TestAwsFindRoleWorksWithOneLoginAssertions(t *testing.T) {
	resp := saml.Response{}
	// For some reason, this is reversed in OneLogin.
	resp.AddAttribute("https://aws.amazon.com/SAML/Attributes/Role", "arn:cloud:iam::1234:role/Admin,arn:cloud:iam::1234:saml-provider/Onelogin")
	pair, _, err := FindRole("", &resp)
	require.True(t, err)
	require.Equal(t, "arn:cloud:iam::1234:saml-provider/Onelogin", pair.ProviderARN)
	require.Equal(t, "arn:cloud:iam::1234:role/Admin", pair.RoleARN)
}
