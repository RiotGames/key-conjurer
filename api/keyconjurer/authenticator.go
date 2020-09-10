package keyconjurer

import (
	"errors"

	"github.com/riotgames/key-conjurer/api/authenticators"
	oneloginduo "github.com/riotgames/key-conjurer/api/authenticators/onelogin_duo"
	"github.com/riotgames/key-conjurer/api/consts"
	"github.com/riotgames/key-conjurer/api/settings"

	"github.com/sirupsen/logrus"
)

func newAuthenticator(logger *logrus.Entry, keyConjurerSettings *settings.Settings) authenticators.Authenticator {
	var authenticator authenticators.Authenticator

	switch consts.AuthenticatorSelect {
	case "onelogin":
		logger.Info("using onelogin authenticator")
		authenticator = oneloginduo.New(logger, keyConjurerSettings)
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
