package oneloginduo

import (
	"fmt"
	"strings"

	"github.com/riotgames/key-conjurer/api/authenticators"
	"github.com/riotgames/key-conjurer/api/authenticators/duo"
	"github.com/riotgames/key-conjurer/api/settings"

	saml "github.com/RobotsAndPencils/go-saml"
	"github.com/rnikoopour/onelogin"
)

type OneLoginAuthenticator struct {
	settings *settings.Settings
	mfa      duo.Duo
}

func New(settings *settings.Settings, mfa duo.Duo) authenticators.Authenticator {
	return &OneLoginAuthenticator{
		settings: settings,
		mfa:      mfa,
	}
}

func (ola *OneLoginAuthenticator) Authenticate(username string, password string) ([]authenticators.Account, error) {
	oneLoginClient := NewOneLogin(ola.settings)
	authenticatedUser, err := oneLoginClient.AuthenticateUser(username, password)
	if err != nil {
		return nil, err
	}

	allUserApps, err := oneLoginClient.GetUserApps(authenticatedUser)
	if err != nil {
		return nil, err
	}

	accounts := make([]authenticators.Account, len(allUserApps))
	for index, app := range allUserApps {
		accounts[index] = app
	}

	return accounts, nil
}

func (ola *OneLoginAuthenticator) Authorize(username string, password string, appID string) (*saml.Response, error) {
	oneLoginClient := NewOneLogin(ola.settings)
	stateTokenResponse, err := oneLoginClient.GetStateToken(username, password, appID)
	if err != nil {
		return nil, err
	}

	device := &onelogin.Device{}
	for i, aDevice := range stateTokenResponse.Devices {
		if aDevice.DeviceType == "Duo Duo Security" {
			device = &stateTokenResponse.Devices[i]
		}
	}
	signatures := strings.Split(device.SignatureRequest, ":")
	txSignature := signatures[0]
	appSignature := signatures[1]

	mfaCookie, err := ola.mfa.SendPush(txSignature, stateTokenResponse.StateToken, stateTokenResponse.CallbackUrl, device.ApiHostName)
	if err != nil {
		return nil, err
	}

	mfaToken := fmt.Sprintf("%v:%v", mfaCookie, appSignature)
	samlString, err := oneLoginClient.GetSamlAssertion(mfaToken, stateTokenResponse.StateToken, appID, fmt.Sprint(device.Id))
	if err != nil {
		return nil, err
	}

	return saml.ParseEncodedResponse(samlString)
}
