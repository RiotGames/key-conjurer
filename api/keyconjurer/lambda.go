package keyconjurer

import (
	"fmt"
	log "keyconjurer-lambda/logger"
	"time"
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
	auth := newAuthenticator()

	logger := log.NewLogger("KeyConjurer-Lambda", event.Client, event.ClientVersion, log.DEBUG)

	// make new keyconjurer instance
	client := NewKeyConjurer(event.Client, event.ClientVersion, auth, logger)

	// get username:password and decrypt if necessary
	user := NewUser(event.Username, event.Password)
	if event.Username == "encrypted" {
		if err := client.AWSClient.Decrypt(event.Password, user); err != nil {
			return CreateResponseError("Invalid username or password"), nil
		}
	}

	// create new logger
	client.Logger.SetUsername(user.Username)

	// authenticate user and get UserData
	userData, err := client.GetUserData(user)
	if err != nil {
		return CreateResponseError("Invalid username or password"), nil
	}

	if event.ShouldEncryptCreds && event.Username != "encrypted" {
		ciphertext, err := client.AWSClient.Encrypt(user)

		if err != nil {
			return CreateResponseUnexpectedError(), nil
		}

		userData.SetCreds(ciphertext)
	} else {
		userData.SetCreds("")
	}

	client.Logger.Info("Successfully retrieved user data")
	return CreateResponseSuccess(userData), nil
}

///////////////////////////////////////////////////////////
//
//    CREDENTIALS
//
///////////////////////////////////////////////////////////

// Event holds incoming data from the user
//  AppID is the OneLogin AppID
type GetAWSCredsEvent struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	AppID          string `json:"appId"`
	Client         string `json:"client"`
	ClientVersion  string `json:"clientVersion"`
	TimeoutInHours int    `json:"timeoutInHours"`
}

// getAwsCreds authenticates the user against OneLogin, then sends a Duo push request to
//  the user, validates the MFA response with OneLogin, then generates STS credentials
//  for the user
func GetAWSCredsEventHandler(event GetAWSCredsEvent) (*Response, error) {
	auth := newAuthenticator()

	logger := log.NewLogger("KeyConjurer-Lambda", event.Client, event.ClientVersion, log.DEBUG)

	// make new keyconjurer instance
	client := NewKeyConjurer(event.Client, event.ClientVersion, auth, logger)

	user := NewUser(event.Username, event.Password)
	if event.Username == "encrypted" {
		if err := client.AWSClient.Decrypt(event.Password, user); err != nil {
			client.Logger.Info("Creds decryption failure", err.Error())
			return CreateResponseError("Invalid username or password"), nil
		}
	}
	client.Logger.SetUsername(user.Username)

	credentials, err := client.GetAwsCreds(user, event.AppID, event.TimeoutInHours)
	if err != nil {
		client.Logger.Info("Key failure", err.Error())
		return CreateResponseError("Unable to get aws credentials"), nil
	}

	client.Logger.Info(fmt.Sprintf("AccessKeyId: %v", *credentials.AccessKeyId), "Key Success")

	stsToken := STSTokenResponse{
		AccessKeyID:     credentials.AccessKeyId,
		SecretAccessKey: credentials.SecretAccessKey,
		SessionToken:    credentials.SessionToken,
		Expiration:      credentials.Expiration.Format(time.RFC3339),
	}

	return CreateResponseSuccess(stsToken), nil
}
