package keyconjurer

import (
	"github.com/riotgames/key-conjurer/api/core"
)

const (
	// AuthenticationProviderDefault lets the server decide which authentication provider to use.  This is not recommended.
	// Older clients will supply this as it has the value of an empty string.
	AuthenticationProviderDefault  AuthenticationProviderName = ""
	AuthenticationProviderOkta     AuthenticationProviderName = "okta"
	AuthenticationProviderOneLogin AuthenticationProviderName = "onelogin"
)

type AuthenticationProviderName = string

type providerMap map[AuthenticationProviderName]core.AuthenticationProvider

func (m *providerMap) Get(name AuthenticationProviderName) (core.AuthenticationProvider, bool) {
	if name == AuthenticationProviderDefault {
		name = AuthenticationProviderOneLogin
	}

	p, ok := (*m)[name]
	return p, ok
}
