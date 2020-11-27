package oneloginduo

import (
	"context"
	"strconv"
	"strings"

	"github.com/riotgames/key-conjurer/api/authenticators"
	"github.com/riotgames/key-conjurer/api/settings"

	"github.com/rnikoopour/onelogin"
)

// OneLogin is a wrapper around the onelogin library
type OneLogin struct {
	ReadUserClient *onelogin.Client
	SamlClient     *onelogin.Client
}

type OneLoginApp struct {
	app onelogin.App
}

func (app OneLoginApp) ID() string {
	return strconv.FormatInt(app.app.ID, 10)
}

func (app OneLoginApp) Name() string {
	return app.app.Name
}

var _ authenticators.Account = &OneLoginApp{}

// NewOneLogin creates a new onelogin client using the providing settings
//  and logs with provided logger
func NewOneLogin(settings *settings.Settings) *OneLogin {
	var readUserClient = onelogin.New(settings.OneLoginReadUserID, settings.OneLoginReadUserSecret,
		settings.OneLoginShard, settings.OneLoginSubdomain)
	var samlClient = onelogin.New(settings.OneLoginSamlID, settings.OneLoginSamlSecret,
		settings.OneLoginShard, settings.OneLoginSubdomain)
	return &OneLogin{
		ReadUserClient: readUserClient,
		SamlClient:     samlClient,
	}
}

// AuthenticateUser validates the user against OneLogin
func (o *OneLogin) AuthenticateUser(username, password string) (*onelogin.AuthenticatedUser, error) {
	return o.SamlClient.Oauth.Authenticate(context.Background(), username, password)
}

// GetUserApps returns the list of OneLogin apps the user has access to
func (o *OneLogin) GetUserApps(user *onelogin.AuthenticatedUser) ([]OneLoginApp, error) {
	oneloginApps, err := o.ReadUserClient.User.GetApps(context.Background(), user.ID)
	if err != nil {
		return nil, err
	}

	convertedApps := []OneLoginApp{}
	for _, app := range *oneloginApps {
		if strings.HasPrefix(app.Name, "AWS") {
			convertedApps = append(convertedApps, OneLoginApp{app: app})
		}
	}

	return convertedApps, nil
}

// GetStateToken retrieves the token necessary to perform MFA with Duo
func (o *OneLogin) GetStateToken(username, password, appID string) (*onelogin.MFAResponse, error) {
	return o.SamlClient.SAML.SamlAssertion(context.Background(), username, password, appID)
}

// GetSamlAssertion retrieve the SAML assertion from OneLogin after MFA happens
func (o *OneLogin) GetSamlAssertion(mfaToken, stateToken, appID, deviceID string) (string, error) {
	samlAssertion, err := o.SamlClient.SAML.VerifyFactor(context.Background(), mfaToken, stateToken, appID, deviceID)
	if err != nil {
		return "", ErrorUnableToGetSamlAssertion
	}

	return samlAssertion, nil
}
