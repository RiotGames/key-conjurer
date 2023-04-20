package okta

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/providers"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

type Authenticator struct {
	client         *okta.Client
	oktaAuthClient AuthClient
	// Storing context in a struct is usually a bad idea, but the Okta SDK gives us one
	ctx context.Context
}

// Authenticate determines whether or not the users current credentials are valid.
func (a Authenticator) Authenticate(ctx context.Context, creds providers.Credentials) (core.User, error) {
	req := AuthRequest{Username: creds.Username, Password: creds.Password}
	res, err := a.oktaAuthClient.VerifyCredentials(ctx, req)
	if err != nil {
		return core.User{}, err
	}

	id, err := res.FindUserID()
	if err != nil {
		return core.User{}, err
	}

	return core.User{ID: id}, err
}

// ListApplications should list all the applications the given user is entitled to access.
func (a Authenticator) ListApplications(ctx context.Context, user core.User) ([]core.Application, error) {
	links, resp, err := a.client.User.ListAppLinks(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	for resp.HasNextPage() {
		var next []*okta.AppLink
		if resp, err = resp.Next(a.ctx, &next); err != nil {
			return nil, err
		}

		links = append(links, next...)
	}

	var cloudAccounts []core.Application
	for _, app := range links {
		// TODO: It seems like a bad idea to put this filtering here as it inherently makes it an AWS-okta provider, rather than just an Okta one.
		// We use the app links endpoint because it's the easiest way to find the applications a user may access.
		if app.AppName != "amazon_aws" && !strings.Contains(app.AppName, "tencent") {
			continue
		}

		cloudAccounts = append(cloudAccounts, core.Application{
			LegacyID: 0,
			ID:       app.AppInstanceId,
			Name:     app.Label,
			Href:     app.LinkUrl,
		})
	}

	return cloudAccounts, nil
}

// GenerateSAMLAssertion should generate a SAML assertion that the user may exchange with the target application in order to gain access to it.
// This will initiate a multi-factor request with Duo.
func (a Authenticator) GenerateSAMLAssertion(ctx context.Context, creds providers.Credentials, appID string) (*core.SAMLResponse, error) {
	if appID == "" {
		return nil, fmt.Errorf("%w: appID cannot be an empty string", core.ErrBadRequest)
	}

	// Verify that the application exists
	// This prevents the user going through the full authentication flow when submitting an invalid application ID.
	// TODO: This step can be skipped by capturing the 'href' from get_user_data and having the user provide it to us, instead of an application ID;
	// this would enable this entire flow to be moved to the client.
	var app okta.Application
	if _, _, err := a.client.Application.GetApplication(ctx, appID, &app, query.NewQueryParams()); err != nil {
		return nil, core.WrapError(core.ErrApplicationNotFound, err)
	}

	href, ok := getHrefLink(app)
	if !ok {
		return nil, fmt.Errorf("could not find application link for %s for user %s", appID, creds.Username)
	}

	source := ApplicationSAMLSource(href)
	assertionBytes, err := source.GetAssertion(ctx, creds.Username, creds.Password)
	if err != nil {
		return nil, fmt.Errorf("could not generate assertion: %w", err)
	}

	assert, err := core.ParseEncodedResponse(string(assertionBytes))
	if err != nil {
		return nil, fmt.Errorf("could not parse assertion: %w", err)
	}

	return assert, nil
}

// New creates a new Okta authenticator.
// An error may be returned if the token or host are in the incorrect format - please refer to the Okta documentation at github.com/okta/okta-sdk-golang/
func New(host, token string) (Authenticator, error) {
	// This is a bit of a hack, but if we assume that the URL passed will always a hostname, we must add https:// ourselves
	// If we do not add https, Okta will complain, and if we do, our own code will break
	uri := url.URL{Host: host, Scheme: "https"}
	ctx, client, err := okta.NewClient(context.Background(), okta.WithOrgUrl(uri.String()), okta.WithToken(token))
	if err != nil {
		return Authenticator{}, err
	}

	return Authenticator{client: client, ctx: ctx, oktaAuthClient: NewAuthClient(host)}, nil
}

// translateOktaError converts an error from Okta to one of the standard provider's errors.
// If the function can't translate the error, it returns a specified default error.
func translateOktaError(err error, defaultErr error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrOktaBadRequest):
		return core.ErrBadRequest
	case errors.Is(err, ErrOktaUnauthorized):
		return core.ErrAuthenticationFailed
	case errors.Is(err, ErrOktaForbidden):
		return core.ErrAccessDenied
	case errors.Is(err, ErrOktaInternalServerError):
		return core.ErrInternalError
	default:
		return defaultErr
	}
}

// wrapOktaError wraps an error from Okta into a standard authentication provider error.
func wrapOktaError(err error, defaultCoreErr error) error {
	return core.WrapError(translateOktaError(err, defaultCoreErr), err)
}

var _ providers.Provider = &Authenticator{}
