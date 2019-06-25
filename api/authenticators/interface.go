package authenticators

import (
	saml "github.com/RobotsAndPencils/go-saml"
)

type Account interface {
	ID() int64
	Name() string
}

// Authenticator is a interface to provide both authentication and authorization
// for getting STS tokens
type Authenticator interface {
	SetMFA(MFA)
	// Authenticate should only authenticate and provide back the information about
	// which accounts/apps a user may be authorized to access
	// i.e. check credentials against identity provider
	Authenticate(username string, password string) ([]Account, error)
	// Authorize should generate as SAML assertion based on the results
	// of whatever authorization mechanisms
	//  i.e. do authz check, MFA, and return a SAML assertion
	Authorize(username string, password string, appID string) (SamlResponse, error)
}

type SamlResponse interface {
	GetBase64String() string
	GetSamlResponse() *saml.Response
}

type MFA interface {
	Do(...string) (string, error)
}
