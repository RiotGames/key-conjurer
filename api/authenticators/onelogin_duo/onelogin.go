package oneloginduo

import (
	"context"

	"github.com/riotgames/key-conjurer/api/settings"

	"github.com/rnikoopour/onelogin"
)

// OneLogin is a wrapper around the onelogin library
type OneLogin struct {
	ReadUserClient *onelogin.Client
	SamlClient     *onelogin.Client
}

// New creates a new onelogin client using the providing settings
func newOneLogin(settings *settings.Settings) *OneLogin {
	var readUserClient = onelogin.New(settings.OneLoginReadUserID, settings.OneLoginReadUserSecret,
		settings.OneLoginShard, settings.OneLoginSubdomain)
	var samlClient = onelogin.New(settings.OneLoginSamlID, settings.OneLoginSamlSecret,
		settings.OneLoginShard, settings.OneLoginSubdomain)
	return &OneLogin{
		ReadUserClient: readUserClient,
		SamlClient:     samlClient,
	}
}

// GetSamlAssertion retrieve the SAML assertion from OneLogin after MFA happens
func (o *OneLogin) GetSamlAssertion(ctx context.Context, mfaToken, stateToken, appID, deviceID string) (string, error) {
	samlAssertion, err := o.SamlClient.SAML.VerifyFactor(ctx, mfaToken, stateToken, appID, deviceID)
	if err != nil {
		return "", ErrorUnableToGetSamlAssertion
	}

	return samlAssertion, nil
}
