package oneloginduo

import (
	"testing"

	"keyconjurer-lambda/authenticators"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticatorOneLoginDuo(t *testing.T) {
	var auth authenticators.Authenticator
	auth = OneLoginAuthenticator{}
	authv, ok := auth.(OneLoginAuthenticator)

	t.Logf("%#v %#v\n", authv, ok)

	assert.EqualValues(t, true, ok, "one login authenticator should comply")

	var account authenticators.Account
	account = OneLoginApp{}
	accountv, ok := account.(OneLoginApp)

	t.Logf("%#v %#v\n", accountv, ok)

	assert.EqualValues(t, true, ok, "one login app should comply")

	var mfa authenticators.MFA
	mfa = DuoMFA{}
	mfav, ok := mfa.(DuoMFA)

	t.Logf("%#v %#v\n", mfav, ok)

	assert.EqualValues(t, true, ok, "one login authenticator should comply")
}
