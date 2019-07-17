package oneloginduo

import (
	"testing"

	"keyconjurer-lambda/authenticators"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticatorOneLoginDuo(t *testing.T) {
	// These are structured this way to ensure that the implemented interfaces
	//  are properly met
	var auth authenticators.Authenticator
	auth = &OneLoginAuthenticator{}
	_, ok := auth.(*OneLoginAuthenticator)
	assert.EqualValues(t, true, ok, "one login authenticator should comply")

	var account authenticators.Account
	account = OneLoginApp{}
	_, ok2 := account.(OneLoginApp)
	assert.EqualValues(t, true, ok2, "one login app should comply")

	var mfa authenticators.MFA
	mfa = &DuoMFA{}
	_, ok3 := mfa.(*DuoMFA)
	assert.EqualValues(t, true, ok3, "one login authenticator should comply")
}
