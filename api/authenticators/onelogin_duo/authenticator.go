package oneloginduo

import (
	"fmt"
	"os"
	"strings"

	"encoding/base64"
	"encoding/json"

	"keyconjurer-lambda/authenticators"
	"keyconjurer-lambda/consts"
	log "keyconjurer-lambda/logger"

	saml "github.com/RobotsAndPencils/go-saml"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/rnikoopour/onelogin"
)

type OneLoginSaml struct {
	b64String    string
	samlResponse *saml.Response
}

func (s OneLoginSaml) GetBase64String() string {
	return s.b64String
}

func (s OneLoginSaml) GetSamlResponse() *saml.Response {
	return s.samlResponse
}

type DuoMFA struct{}

func (d DuoMFA) Do(args ...string) (string, error) {
	logger := log.NewLogger("KeyConjurer", "DuoMFA",
		fmt.Sprintf("%s-%s", consts.Version, "duo"), log.DEBUG)

	logger.Info("KeyConjurer", "authenticator", "DuoMFA", "creating new duo MFA")
	duo := NewDuo(logger)

	if len(args) < 4 {
		return "", ErrorDuoArgsError
	}

	txSig := args[0]
	stateToken := args[1]
	callbackUrl := args[2]
	apiHost := args[3]

	logger.Info("KeyConjurer", "authenticator", "DuoMFA", "sending duo push")
	return duo.SendPush(txSig, stateToken, callbackUrl, apiHost)
}

type OneLoginAuthenticator struct {
	Settings *Settings
	MFA      authenticators.MFA
}

func New() authenticators.Authenticator {
	logger := log.NewLogger("KeyConjurer", "authenticator",
		fmt.Sprintf("%s-%s", consts.Version, "onelogin"), log.DEBUG)

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

	//onelogin := NewOneLogin(settings, logger)
	//duo := NewDuo(logger)

	return &OneLoginAuthenticator{Settings: settings}
}

func (ola *OneLoginAuthenticator) SetMFA(mfa authenticators.MFA) {
	ola.MFA = mfa
}

func (ola *OneLoginAuthenticator) Authenticate(username string, password string) ([]authenticators.Account, error) {
	logger := log.NewLogger("KeyConjurer", "authenticator",
		fmt.Sprintf("%s-%s", consts.Version, "onelongin"), log.DEBUG)

	oneLoginClient := NewOneLogin(ola.Settings, logger)

	authenticatedUser, err := oneLoginClient.AuthenticateUser(username, password)
	if err != nil {
		logger.Error("KeyConjurer", "onelogin", "Failed to authenticate user", err.Error())
		return nil, err
	}

	allUserApps, err := oneLoginClient.GetUserApps(authenticatedUser)
	if err != nil {
		logger.Error("KeyConjurer", "onelogin", "Unable to get user apps", err.Error())
		return nil, err
	}

	accounts := make([]authenticators.Account, len(allUserApps))
	for index, app := range allUserApps {
		accounts[index] = app
	}

	return accounts, nil
}

func (ola *OneLoginAuthenticator) Authorize(username string, password string, appID string) (authenticators.SamlResponse, error) {
	logger := log.NewLogger("KeyConjurer", "authenticator",
		fmt.Sprintf("%s-%s", consts.Version, "onelogin"), log.DEBUG)

	oneLoginClient := NewOneLogin(ola.Settings, logger)

	stateTokenResponse, err := oneLoginClient.GetStateToken(username, password, appID)
	if err != nil {
		logger.Error("KeyConjurer", "Authorize", "Unable to get state token", err.Error())
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
		logger.Error("KeyConjurer", "Authorize", "mfa is nil")
	}

	logger.Info("KeyConjurer", "Authorize", "Sending mfa push")
	mfaCookie, err := ola.MFA.Do(txSignature, stateTokenResponse.StateToken, stateTokenResponse.CallbackUrl, device.ApiHostName)
	if err != nil {
		logger.Error("KeyConjurer", "Authorize", "Unable to get mfaCookie", err.Error())
		return nil, err
	}

	mfaToken := fmt.Sprintf("%v:%v", mfaCookie, appSignature)
	logger.Info("KeyConjurer", "Authorize", "Getting SAML assertion")
	samlString, err := oneLoginClient.GetSamlAssertion(mfaToken, stateTokenResponse.StateToken, appID, fmt.Sprint(device.Id))
	if err != nil {
		logger.Error("KeyConjurer", "Authorize", "Unable to get SAML Assertion")
		return nil, err
	}

	response, err := saml.ParseEncodedResponse(samlString)
	if err != nil {
		logger.Error("KeyConjurer", "Authorize", "Unable to parse SAML Assertion into SAML Response")
		return nil, err
	}
	return OneLoginSaml{b64String: samlString, samlResponse: response}, nil
}
