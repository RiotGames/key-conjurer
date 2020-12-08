package keyconjurer

import (
	"errors"

	"github.com/riotgames/key-conjurer/api/authenticators"
	oneloginduo "github.com/riotgames/key-conjurer/api/authenticators/onelogin_duo"
	"github.com/riotgames/key-conjurer/api/consts"
	"github.com/riotgames/key-conjurer/api/settings"

	"github.com/sirupsen/logrus"
)

var errNoAuthenticatorSelected = errors.New("no authenticator selected")

func newAuthenticator(logger *logrus.Entry, keyConjurerSettings *settings.Settings) (authenticators.Authenticator, error) {
	var authenticator authenticators.Authenticator
	switch consts.AuthenticatorSelect {
	case "onelogin":
		logger.Info("using onelogin authenticator")
		authenticator = oneloginduo.New(logger, keyConjurerSettings)
	default:
		return nil, errNoAuthenticatorSelected
	}

	switch consts.MFASelect {
	case "duo":
		logger.Info("using duo mfa")
		duo := oneloginduo.NewDuoMFA(logger)
		authenticator.SetMFA(duo)
	default:
		return nil, errNoAuthenticatorSelected
	}

	return authenticator, nil
}
