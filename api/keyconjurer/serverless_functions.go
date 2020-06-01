package keyconjurer

import (
	"errors"

	log "keyconjurer-lambda/logger"
	"keyconjurer-lambda/settings"

	"github.com/sirupsen/logrus"
)

///////////////////////////////////////////////////////////
//
//    USER DATA
//
///////////////////////////////////////////////////////////

// Event holds incoming data from the user
//  ShouldEncryptCreds determines if the user receieves encrypted credentials in their response
type GetUserDataEvent struct {
	Username           string `json:"username"`
	Password           string `json:"password"`
	Client             string `json:"client"`
	ClientVersion      string `json:"clientVersion"`
	ShouldEncryptCreds bool   `json:"shouldEncryptCreds"`
}

// GetUserData authenticates the user against OneLogin and retrieves a list of
//  AWS application the user has available
func GetUserDataEventHandler(event GetUserDataEvent) (*Response, error) {
	logger := log.NewLogger(event.Client, event.ClientVersion, logrus.DebugLevel)
	keyConjurerSettings := settings.NewSettings(logger)

	auth := newAuthenticator(logger, keyConjurerSettings)

	// make new keyconjurer instance
	client := NewKeyConjurer(event.Client, event.ClientVersion, auth, logger, keyConjurerSettings)

	// get username:password and decrypt if necessary

	user, err := client.providerClient.GetUserCredentials(event.Username, event.Password)
	if err != nil {
		return CreateResponseError(err.Error()), errors.New("unable to get user information with current provider")
	}

	// Set the username field permanently for future logs
	client.Logger = client.Logger.WithFields(logrus.Fields{
		"username": user.Username})

	// authenticate user and get UserData
	userData, err := client.GetUserData(user)
	if err != nil {
		return CreateResponseError("Invalid username or password"), nil
	}

	if event.ShouldEncryptCreds && event.Username != "encrypted" {
		ciphertext, err := client.providerClient.EncryptUserInformation(user)
		if err != nil {
			return CreateResponseUnexpectedError(), nil
		}

		userData.SetCreds(ciphertext)
	} else {
		userData.SetCreds("")
	}

	logger.Info("successfully retrieved user data")
	return CreateResponseSuccess(userData), nil
}

///////////////////////////////////////////////////////////
//
//    CREDENTIALS
//
///////////////////////////////////////////////////////////

// Event holds incoming data from the user
//  AppID is the OneLogin AppID
type GetTemporaryCredentialEvent struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	AppID          string `json:"appId"`
	Client         string `json:"client"`
	ClientVersion  string `json:"clientVersion"`
	TimeoutInHours int    `json:"timeoutInHours"`
}

func GetTemporaryCredentialEventHandler(event GetTemporaryCredentialEvent) (*Response, error) {
	logger := log.NewLogger(event.Client, event.ClientVersion, logrus.DebugLevel)
	keyConjurerSettings := settings.NewSettings(logger)

	auth := newAuthenticator(logger, keyConjurerSettings)

	// make new keyconjurer instance
	client := NewKeyConjurer(event.Client, event.ClientVersion, auth, logger, keyConjurerSettings)

	user, err := client.providerClient.GetUserCredentials(event.Username, event.Password)
	if err != nil {
		return CreateResponseError(err.Error()), errors.New("unable to get user information with current provider")
	}

	// Set the username field permanently for future logs
	client.Logger = client.Logger.WithFields(logrus.Fields{
		"username": user.Username})

	credentials, err := client.GetTemporaryCredentialsForUser(user, event.AppID, event.TimeoutInHours)
	if err != nil {
		client.Logger.Info("Key failure", err.Error())
		return CreateResponseError("unable to get aws credentials"), nil
	}

	return CreateResponseSuccess(credentials), nil
}
