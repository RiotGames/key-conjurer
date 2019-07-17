package oneloginduo

import (
	"fmt"
	"os"
	"strings"

	"encoding/base64"
	"encoding/json"

	"keyconjurer-lambda/authenticators"

	saml "github.com/RobotsAndPencils/go-saml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
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
	d.logger.Info("KeyConjurer", "authenticator", "DuoMFA", "creating new duo MFA")
	duo := NewDuo(d.logger)

	if len(args) < 4 {
		return "", ErrorDuoArgsError
	}

	txSig := args[0]
	stateToken := args[1]
	callbackUrl := args[2]
	apiHost := args[3]

	d.logger.Info("KeyConjurer", "authenticator", "DuoMFA", "sending duo push")
	return duo.SendPush(txSig, stateToken, callbackUrl, apiHost)
}

type OneLoginAuthenticator struct {
	Settings *Settings
	MFA      authenticators.MFA
	logger   *logrus.Entry
}

func New(logger *logrus.Entry) authenticators.Authenticator {
	awsRegion := os.Getenv("AWSRegion")
	settings := &Settings{AwsRegion: awsRegion}

	awsConfig := &aws.Config{
		Region: aws.String("us-west-2"),
	}

	awsSession := session.Must(session.NewSession(awsConfig))

	kmsSession := kms.New(awsSession)

	encryptedSettings := os.Getenv("EncryptedSettings")

	blob, err := base64.StdEncoding.DecodeString(encryptedSettings)
	if err != nil {
		logger.Error("KeyConjurer", "AWSClient", "Unable to decode ciphertext", err.Error())
		// should handle the
		panic(err)
	}

	input := &kms.DecryptInput{
		CiphertextBlob: blob,
	}

	result, err := kmsSession.Decrypt(input)
	if err != nil {
		logger.Error("KeyConjurer", "AWSClient", "authenticator Failed to decrypt", err.Error())
		panic(err)
	}

	if err := json.Unmarshal(result.Plaintext, settings); err != nil {
		logger.Error("KeyConjurer", "AWSClient", "Unable to unmarshal", err.Error())
		panic(err)
	}

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
		ola.logger.Error("KeyConjurer", "onelogin", "Failed to authenticate user", err.Error())
		return nil, err
	}

	allUserApps, err := oneLoginClient.GetUserApps(authenticatedUser)
	if err != nil {
		ola.logger.Error("KeyConjurer", "onelogin", "Unable to get user apps", err.Error())
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
		ola.logger.Error("KeyConjurer", "Authorize", "Unable to get state token", err.Error())
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
		ola.logger.Error("KeyConjurer", "Authorize", "mfa is nil")
	}

	ola.logger.Info("KeyConjurer", "Authorize", "Sending mfa push")
	mfaCookie, err := ola.MFA.Do(txSignature, stateTokenResponse.StateToken, stateTokenResponse.CallbackUrl, device.ApiHostName)
	if err != nil {
		ola.logger.Error("KeyConjurer", "Authorize", "Unable to get mfaCookie", err.Error())
		return nil, err
	}

	mfaToken := fmt.Sprintf("%v:%v", mfaCookie, appSignature)
	ola.logger.Info("KeyConjurer", "Authorize", "Getting SAML assertion")
	samlString, err := oneLoginClient.GetSamlAssertion(mfaToken, stateTokenResponse.StateToken, appID, fmt.Sprint(device.Id))
	if err != nil {
		ola.logger.Error("KeyConjurer", "Authorize", "Unable to get SAML Assertion")
		return nil, err
	}

	response, err := saml.ParseEncodedResponse(samlString)
	if err != nil {
		ola.logger.Error("KeyConjurer", "Authorize", "Unable to parse SAML Assertion into SAML Response")
		return nil, err
	}
	return OneLoginSaml{
		b64String:    samlString,
		samlResponse: response,
		logger:       ola.logger}, nil
}
