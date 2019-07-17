package oneloginduo

import (
	"context"
	"fmt"
	"strings"

	"github.com/rnikoopour/onelogin"
	"github.com/sirupsen/logrus"
)

// OneLogin is a wrapper around the onelogin library
type OneLogin struct {
	ReadUserClient *onelogin.Client
	SamlClient     *onelogin.Client
	logger         *logrus.Entry
}

type OneLoginApp struct {
	app onelogin.App
}

func (app OneLoginApp) ID() int64 {
	return app.app.ID
}

func (app OneLoginApp) Name() string {
	return app.app.Name
}

// NewOneLogin creates a new onelogin client using the providing settings
//  and logs with provided logger
func NewOneLogin(settings *Settings, logger *logrus.Entry) *OneLogin {
	var readUserClient = onelogin.New(settings.OneLoginReadUserID, settings.OneLoginReadUserSecret,
		settings.OneLoginShard, settings.OneLoginSubdomain)
	var samlClient = onelogin.New(settings.OneLoginSamlID, settings.OneLoginSamlSecret,
		settings.OneLoginShard, settings.OneLoginSubdomain)
	return &OneLogin{
		ReadUserClient: readUserClient,
		SamlClient:     samlClient,
		logger:         logger}
}

// AuthenticateUser validates the user against OneLogin
func (o *OneLogin) AuthenticateUser(username, password string) (*onelogin.AuthenticatedUser, error) {
	user, err := o.SamlClient.Oauth.Authenticate(context.Background(), username, password)
	if err != nil {
		o.logger.Error("OneLogin", fmt.Sprintf("Unable to authenticate %v", username), err.Error())
		return nil, err
	}
	return user, nil
}

// GetUserApps returns the list of OneLogin apps the user has access to
func (o *OneLogin) GetUserApps(user *onelogin.AuthenticatedUser) ([]OneLoginApp, error) {
	oneloginApps, err := o.ReadUserClient.User.GetApps(context.Background(), user.ID)
	if err != nil {
		o.logger.Error("OneLogin", "Unable to get user apps", err.Error())
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
	response, err := o.SamlClient.SAML.SamlAssertion(context.Background(), username, password, appID)
	if err != nil {
		o.logger.Error("OneLogin", "Unable to get state token", err.Error())
		return nil, err
	}
	return response, nil
}

// GetSamlAssertion retrieve the SAML assertion from OneLogin after MFA happens
func (o *OneLogin) GetSamlAssertion(mfaToken, stateToken, appID, deviceID string) (string, error) {
	samlAssertion, err := o.SamlClient.SAML.VerifyFactor(context.Background(), mfaToken, stateToken, appID, deviceID)
	if err != nil {
		o.logger.Error("OneLogin", "Unable to get SAML assertion", err.Error())
		return "", ErrorUnableToGetSamlAssertion
	}

	o.logger.Info("onelogin_duo", "saml assertion", samlAssertion)

	return samlAssertion, nil
}
