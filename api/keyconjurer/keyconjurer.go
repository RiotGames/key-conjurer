package keyconjurer

import (
	"os"

	"keyconjurer-lambda/authenticators"
	"keyconjurer-lambda/keyconjurer/awsclient"
	"keyconjurer-lambda/keyconjurer/settings"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/sirupsen/logrus"
)

// KeyConjurer is used to generate temporary AWS credentials
type KeyConjurer struct {
	AWSClient     *awsclient.AWSClient
	Authenticator authenticators.Authenticator
	Logger        *logrus.Entry
}

// New creates an KeyConjurer service
func NewKeyConjurer(client, clientVersion string, auth authenticators.Authenticator, logger *logrus.Entry, keyConjurerSettings *settings.Settings) *KeyConjurer {
	awsRegion := os.Getenv("AWSRegion")
	awsClient := awsclient.NewAWSClient(awsRegion, logger)
	awsClient.SetKMSKeyID(keyConjurerSettings.AwsKMSKeyID)

	return &KeyConjurer{
		AWSClient:     awsClient,
		Authenticator: auth,
		Logger:        logger,
	}
}

// GetUserData retrieves the users devices and apps from OneLogin. The apps
//  are filtered to only include the AWS related applications
func (a *KeyConjurer) GetUserData(user *User) (*UserData, error) {
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
func (a *KeyConjurer) GetAwsCreds(user *User, appID string, keyTimeoutInHours int) (*sts.Credentials, error) {
	samlAssertion, err := a.Authenticator.Authorize(user.Username, user.Password, appID)
	if err != nil {
		a.Logger.Error("unable to parse saml assertion reason: ", err.Error())
		return nil, err
	}

	roleArn, principalArn, err := a.AWSClient.SelectRoleFromSaml(samlAssertion)
	if err != nil {
		a.Logger.Error("unable to select role from saml reason: ", err.Error())
		return nil, err
	}

	a.Logger.Info("KeyConjurer", "Assuming role")
	credentials, err := a.AWSClient.AssumeRole(roleArn, principalArn, samlAssertion, keyTimeoutInHours)
	if err != nil {
		a.Logger.Error("unable to assume role reason: ", err.Error())
		return nil, err
	}
	return credentials, nil
}
