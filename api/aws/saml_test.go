package aws

import (
	"testing"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/stretchr/testify/require"
)

func TestGetRoleDoesntBreakIfYouHaveMultipleRoles(t *testing.T) {
	resp := saml.Response{}
	resp.AddAttribute("https://aws.amazon.com/SAML/Attributes/Role", "arn:aws:iam::1234:saml-provider/Okta,arn:aws:iam::1234:role/Admin")
	resp.AddAttribute("https://aws.amazon.com/SAML/Attributes/Role", "arn:aws:iam::1234:saml-provider/Okta,arn:aws:iam::1234:role/Power")

	providerARN, roleARN, err := getRole("Power", &resp)
	require.NoError(t, err)
	require.Equal(t, "arn:aws:iam::1234:saml-provider/Okta", providerARN)
	require.Equal(t, "arn:aws:iam::1234:role/Power", roleARN)
	providerARN, roleARN, err = getRole("Admin", &resp)
	require.NoError(t, err)
	require.Equal(t, "arn:aws:iam::1234:saml-provider/Okta", providerARN)
	require.Equal(t, "arn:aws:iam::1234:role/Admin", roleARN)
}

func TestGetRoleWorksWithOneLoginAssertions(t *testing.T) {
	resp := saml.Response{}
	// For some reason, this is reversed in OneLogin.
	resp.AddAttribute("https://aws.amazon.com/SAML/Attributes/Role", "arn:aws:iam::1234:role/Admin,arn:aws:iam::1234:saml-provider/Onelogin")
	providerARN, roleARN, err := getRole("", &resp)
	require.NoError(t, err)
	require.Equal(t, "arn:aws:iam::1234:saml-provider/Onelogin", providerARN)
	require.Equal(t, "arn:aws:iam::1234:role/Admin", roleARN)
}
