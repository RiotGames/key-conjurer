package keyconjurer

import (
	"errors"
	"keyconjurer-lambda/authenticators"
	oneloginduo "keyconjurer-lambda/authenticators/onelogin_duo"
	"keyconjurer-lambda/consts"

	"github.com/sirupsen/logrus"
)

func newAuthenticator(logger *logrus.Entry) authenticators.Authenticator {
	var authenticator authenticators.Authenticator

	switch consts.AuthenticatorSelect {
	case "onelogin":
		logger.Info("using onelogin authenticator")
		authenticator = oneloginduo.New(logger)
	default:
		panic(errors.New("No Authenticator Selected"))
	}

	switch consts.MFASelect {
	case "duo":
		logger.Info("using duo mfa")
		duo := oneloginduo.NewDuoMFA(logger)
		authenticator.SetMFA(duo)
	default:
		panic(errors.New("No Authenticator Selected"))
	}

	return authenticator
}
