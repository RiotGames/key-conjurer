package providers

import (
	"context"

	"github.com/riotgames/key-conjurer/api/core"
)

const (
	// Default lets the server decide which authentication provider to use.  This is not recommended.
	// Older clients will supply this as it has the value of an empty string.
	Default  = ""
	Okta     = "okta"
	OneLogin = "onelogin"
)

// A Provider is a component which will verify user credentials, list the applications a user is entitled to, the roles the user may assume within that application and generate SAML assertions for federation.
type Provider interface {
	// Authenticate should validate that the provided credentials are correct for a user.
	Authenticate(ctx context.Context, credentials Credentials) (core.User, error)
	// ListApplications should list all the applications the given user is entitled to access.
	ListApplications(ctx context.Context, user core.User) ([]core.Application, error)
	// GenerateSAMLAssertion should generate a SAML assertion that the user may exchange with the target application in order to gain access to it.
	GenerateSAMLAssertion(ctx context.Context, credentials Credentials, appID string) (*core.SAMLResponse, error)
}

type Credentials struct {
	Username string
	Password string
}

var providers map[string]Provider = map[string]Provider{}

func Get(name string) (Provider, bool) {
	if name == Default {
		name = OneLogin
	}

	p, ok := providers[name]
	return p, ok
}

func Register(name string, provider Provider) {
	providers[name] = provider
}

func ForEach(cb func(name string, prov Provider)) {
	for k, v := range providers {
		cb(k, v)
	}
}
