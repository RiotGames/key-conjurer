package keyconjurer

import (
	"errors"
	"keyconjurer-lambda/authenticators"
	oneloginduo "keyconjurer-lambda/authenticators/onelogin_duo"
	"keyconjurer-lambda/consts"
	log "keyconjurer-lambda/logger"
)

func newAuthenticator() authenticators.Authenticator {
	logger := log.NewLogger("keyconjurer", "authenticator-factory", consts.Version, log.DEBUG)
	var authenticator authenticators.Authenticator

	switch consts.AuthenticatorSelect {
	case "onelogin":
		logger.Info("KeyConjurer", "authenticator", "using onelogin authenticator")
		authenticator = oneloginduo.New()
	default:
		panic(errors.New("No Authenticator Selected"))
	}

	switch consts.MFASelect {
	case "duo":
		logger.Info("KeyConjurer", "authenticator", "using duo mfa")
		duo := oneloginduo.DuoMFA{}
		authenticator.SetMFA(duo)
	default:
		panic(errors.New("No Authenticator Selected"))
	}

	return authenticator
}
