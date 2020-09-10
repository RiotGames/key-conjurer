package keyconjurer

import (
	"github.com/riotgames/key-conjurer/api/authenticators"
	cloudprovider "github.com/riotgames/key-conjurer/api/cloud-providers"
	"github.com/riotgames/key-conjurer/api/settings"

	"github.com/sirupsen/logrus"
)

// KeyConjurer is used to generate temporary AWS credentials
type KeyConjurer struct {
	providerClient cloudprovider.Provider
	Authenticator  authenticators.Authenticator
	Logger         *logrus.Entry
}

// New creates an KeyConjurer service
func NewKeyConjurer(client, clientVersion string, auth authenticators.Authenticator, logger *logrus.Entry, keyConjurerSettings *settings.Settings) *KeyConjurer {
	provider, err := cloudprovider.NewProvider(keyConjurerSettings, logger)
	if err != nil {
		panic(err)
	}
	return &KeyConjurer{
		providerClient: provider,
		Authenticator:  auth,
		Logger:         logger,
	}
}

// GetUserData retrieves the users devices and apps from OneLogin. The apps
//  are filtered to only include the AWS related applications
func (a *KeyConjurer) GetUserData(user *cloudprovider.User) (*UserData, error) {
	authAccounts, err := a.Authenticator.Authenticate(user.Username, user.Password)
	if err != nil {
		a.Logger.Error("error authenticating reason: ", err.Error())
		return nil, err
	}

	userData := &UserData{
		Devices: make([]Device, 0),
		Apps:    authAccounts,
		Creds:   user.Password,
	}

	return userData, nil
}

// GetAwsCreds authenticates the user against OneLogin, sends a Duo push request
//  to the user, then retrieves AWS credentials
func (a *KeyConjurer) GetTemporaryCredentialsForUser(user *cloudprovider.User, appID string, keyTimeoutInHours int) (interface{}, error) {
	samlAssertion, err := a.Authenticator.Authorize(user.Username, user.Password, appID)
	if err != nil {
		a.Logger.Error("unable to parse saml assertion reason: ", err.Error())
		return nil, err
	}

	return a.providerClient.GetTemporaryCredentialsForUser(samlAssertion, keyTimeoutInHours)
}
