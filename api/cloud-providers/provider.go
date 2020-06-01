package cloudprovider

import (
	"errors"
	"keyconjurer-lambda/authenticators"
	"keyconjurer-lambda/settings"

	"github.com/sirupsen/logrus"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Provider interface {
	GetUserCredentials(username, password string) (*User, error)
	DecryptUserInforamtion(ciphertext string, user interface{}) error
	EncryptUserInformation(data interface{}) (string, error)
	GetTemporaryCredentialsForUser(samlAssertion authenticators.SamlResponse, ttlInHours int) (interface{}, error)
}

func NewProvider(settings *settings.Settings, logger *logrus.Entry) (Provider, error) {
	if settings.AwsRegion != "" && settings.AwsKMSKeyID != "" {
		return NewAWSProvider(settings, logger)
	}

	return nil, errors.New("no matching credentials safe for settings")
}
