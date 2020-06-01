package oneloginduo

import (
	"fmt"
	"strings"

	"keyconjurer-lambda/authenticators"
	"keyconjurer-lambda/settings"

	saml "github.com/RobotsAndPencils/go-saml"
	"github.com/rnikoopour/onelogin"
	"github.com/sirupsen/logrus"
)

type OneLoginSaml struct {
	b64String    string
	samlResponse *saml.Response
	logger       *logrus.Entry
}

func (s OneLoginSaml) GetBase64String() string {
	return s.b64String
}

func (s OneLoginSaml) GetSamlResponse() *saml.Response {
	return s.samlResponse
}

type DuoMFA struct {
	logger *logrus.Entry
}

func NewDuoMFA(logger *logrus.Entry) *DuoMFA {
	return &DuoMFA{
		logger: logger}
}

func (d *DuoMFA) Do(args ...string) (string, error) {
	d.logger.Debug("prepping push")
	duo := NewDuo(d.logger)

	if len(args) < 4 {
		return "", ErrorDuoArgsError
	}

	txSig := args[0]
	stateToken := args[1]
	callbackUrl := args[2]
	apiHost := args[3]

	d.logger.Info("sending duo push")
	return duo.SendPush(txSig, stateToken, callbackUrl, apiHost)
}

type OneLoginAuthenticator struct {
	Settings *settings.Settings
	MFA      authenticators.MFA
	logger   *logrus.Entry
}

func New(logger *logrus.Entry, settings *settings.Settings) authenticators.Authenticator {
	return &OneLoginAuthenticator{
		Settings: settings,
		logger:   logger}
}

func (ola *OneLoginAuthenticator) SetMFA(mfa authenticators.MFA) {
	ola.MFA = mfa
}

func (ola *OneLoginAuthenticator) Authenticate(username string, password string) ([]authenticators.Account, error) {
	oneLoginClient := NewOneLogin(ola.Settings, ola.logger)

	authenticatedUser, err := oneLoginClient.AuthenticateUser(username, password)
	if err != nil {
		ola.logger.Error("failed to authenticate user reason: ", err.Error())
		return nil, err
	}

	allUserApps, err := oneLoginClient.GetUserApps(authenticatedUser)
	if err != nil {
		ola.logger.Error("unable to get user apps reason: ", err.Error())
		return nil, err
	}

	accounts := make([]authenticators.Account, len(allUserApps))
	for index, app := range allUserApps {
		accounts[index] = app
	}

	return accounts, nil
}

func (ola *OneLoginAuthenticator) Authorize(username string, password string, appID string) (authenticators.SamlResponse, error) {
	oneLoginClient := NewOneLogin(ola.Settings, ola.logger)

	stateTokenResponse, err := oneLoginClient.GetStateToken(username, password, appID)
	if err != nil {
		ola.logger.Error("unable to get state token reason: ", err.Error())
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

	if ola.MFA == nil {
		ola.logger.Error("mfa is nil")
	}

	ola.logger.Info("sending mfa push")
	mfaCookie, err := ola.MFA.Do(txSignature, stateTokenResponse.StateToken, stateTokenResponse.CallbackUrl, device.ApiHostName)
	if err != nil {
		ola.logger.Error("unable to get mfacookie reason: ", err.Error())
		return nil, err
	}

	mfaToken := fmt.Sprintf("%v:%v", mfaCookie, appSignature)
	ola.logger.Info("KeyConjurer", "Authorize", "Getting SAML assertion")
	samlString, err := oneLoginClient.GetSamlAssertion(mfaToken, stateTokenResponse.StateToken, appID, fmt.Sprint(device.Id))
	if err != nil {
		ola.logger.Error("unable to get saml assertion")
		return nil, err
	}

	response, err := saml.ParseEncodedResponse(samlString)
	if err != nil {
		ola.logger.Error("unable to parse saml assertion into saml response")
		return nil, err
	}
	return OneLoginSaml{
		b64String:    samlString,
		samlResponse: response,
		logger:       ola.logger}, nil
}
