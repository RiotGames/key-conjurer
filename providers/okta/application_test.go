package okta

import (
	"context"
	"os"
	"testing"

	"github.com/riotgames/key-conjurer/api/cloud"
	"github.com/riotgames/key-conjurer/api/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This is an integration test which performs real authorization actions against a real endpoint.
func Test_ApplicationSAMLSource_GetAssertion(t *testing.T) {
	if _, ok := os.LookupEnv("CI"); ok {
		t.Skipf("Not running %s in CI", t.Name())
	}

	ctx := context.Background()
	source := ApplicationSAMLSource(os.Getenv("OKTA_APPLICATION_URL"))
	assertionBytes, err := source.GetAssertion(ctx, os.Getenv("USER"), os.Getenv("PASS"))
	require.NoError(t, err)
	assertion, err := core.ParseEncodedResponse(string(assertionBytes))
	require.NoError(t, err)
	provider, _ := cloud.NewProvider("us-west-2", "")
	_, resp2, err := provider.GetTemporaryCredentialsForUser(ctx, os.Getenv("ROLE_NAME"), assertion, 8)
	assert.NoError(t, err, "could not generate temporary credentials: %s", err)

	t.Logf("AccessKeyID: %s", *resp2.AccessKeyID)
	t.Logf("SecretAccessKey: %s", *resp2.SecretAccessKey)
	t.Logf("SessionToken: %s", *resp2.SessionToken)
}
