package cloud

import (
	"testing"

	"github.com/RobotsAndPencils/go-saml"
	"github.com/stretchr/testify/require"
)

func TestAwsGetRoleDoesntBreakIfYouHaveMultipleRoles(t *testing.T) {
	resp := saml.Response{}
	resp.AddAttribute("https://cloud.amazon.com/SAML/Attributes/Role", "arn:cloud:iam::1234:saml-provider/Okta,arn:cloud:iam::1234:role/Admin")
	resp.AddAttribute("https://cloud.amazon.com/SAML/Attributes/Role", "arn:cloud:iam::1234:saml-provider/Okta,arn:cloud:iam::1234:role/Power")
	providerARN, roleARN, _, err := getRole("Power", &resp)
	require.NoError(t, err)
	require.Equal(t, "arn:cloud:iam::1234:saml-provider/Okta", providerARN)
	require.Equal(t, "arn:cloud:iam::1234:role/Power", roleARN)
	providerARN, roleARN, _, err = getRole("Admin", &resp)
	require.NoError(t, err)
	require.Equal(t, "arn:cloud:iam::1234:saml-provider/Okta", providerARN)
	require.Equal(t, "arn:cloud:iam::1234:role/Admin", roleARN)
}

func TestAwsGetRoleWorksWithOneLoginAssertions(t *testing.T) {
	resp := saml.Response{}
	// For some reason, this is reversed in OneLogin.
	resp.AddAttribute("https://cloud.amazon.com/SAML/Attributes/Role", "arn:cloud:iam::1234:role/Admin,arn:cloud:iam::1234:saml-provider/Onelogin")
	providerARN, roleARN, _, err := getRole("", &resp)
	require.NoError(t, err)
	require.Equal(t, "arn:cloud:iam::1234:saml-provider/Onelogin", providerARN)
	require.Equal(t, "arn:cloud:iam::1234:role/Admin", roleARN)
}
