package okta

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/riotgames/key-conjurer/api/core"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/riotgames/key-conjurer/api/authenticators/duo"
)

type Authenticator struct {
	client         *okta.Client
	mfa            duo.Duo
	oktaAuthClient oktaAuthClient
	// Storing context in a struct is usually a bad idea, but the Okta SDK gives us one
	ctx context.Context
}

// Authenticate retrieves a list of applications for user with the given username and password.
//
// This will first attempt to validate if the user has the appropriate credentials before returning applications for that user.
func (a *Authenticator) Authenticate(ctx context.Context, creds core.Credentials) (core.User, core.AuthenticationProviderError) {
	req := authnRequest{Username: creds.Username, Password: creds.Password}
	res, err := a.oktaAuthClient.Authn(ctx, req)
	// We don't need to acknowledge this error because we're using zero values all the way down
	return core.User{ID: res.UserID()}, err
}

// ListApplications should list all the applications the given user is entitled to access.
func (a *Authenticator) ListApplications(ctx context.Context, user core.User) ([]core.Application, core.AuthenticationProviderError) {
	// TODO: It seems like a bad idea to put this filtering here as it inherently makes it an AWS-okta provider, rather than just an Okta one.
	// We use the app links endpoint because it's the easiest way to find the applications a user may access.
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
		if app.AppName != "amazon_aws" && !strings.Contains(app.AppName, "tencent") {
			continue
		}

		cloudAccounts = append(cloudAccounts, core.Application{
			LegacyID: 0,
			ID:       app.AppInstanceId,
			Name:     app.Label,
		})
	}

	return cloudAccounts, nil
}

// GenerateSAMLAssertion should generate a SAML assertion that the user may exchange with the target application in order to gain access to it.
// This will initiate a multi-factor request with Duo.
func (a *Authenticator) GenerateSAMLAssertion(ctx context.Context, creds core.Credentials, appID string) (*core.SAMLResponse, core.AuthenticationProviderError) {
	if appID == "" {
		return nil, fmt.Errorf("%w: appID cannot be an empty string", core.ErrBadRequest)
	}

	app, _, err := a.client.Application.GetApplication(ctx, appID, &okta.Application{}, query.NewQueryParams())
	if err != nil {
		return nil, core.WrapError(core.ErrApplicationNotFound, err)
	}

	appl := app.(*okta.Application)

	st, err := a.oktaAuthClient.Authn(ctx, authnRequest{Username: creds.Username, Password: creds.Password})
	if err != nil {
		return nil, wrapOktaError(err, core.ErrAuthenticationFailed)
	}

	var f *okta.UserFactor
	for _, factor := range st.Factors() {
		if factor.Provider == "DUO" && factor.FactorType == "web" {
			f = &factor
			break
		}
	}

	if f == nil {
		return nil, fmt.Errorf("%w: no Duo web factor found", core.ErrInternalError)
	}

	vf, err := a.oktaAuthClient.VerifyFactor(ctx, st.StateToken, *f)
	if err != nil {
		return nil, wrapOktaError(err, core.ErrFactorVerificationFailed)
	}

	tok, err := a.mfa.SendPush(vf.AuthSignature, vf.StateToken.String(), vf.CallbackURL, vf.Host)
	if err != nil {
		return nil, core.WrapError(core.ErrCouldNotSendMfaPush, err)
	}

	if err = a.oktaAuthClient.SubmitChallengeResponse(ctx, vf, tok); err != nil {
		return nil, wrapOktaError(err, core.ErrSubmitChallengeResponseFailed)
	}

	session, err := a.oktaAuthClient.CreateSession(ctx, vf)
	if err != nil {
		return nil, wrapOktaError(err, core.ErrCouldNotCreateSession)
	}

	samlResponse, err := a.oktaAuthClient.GetSAMLResponse(ctx, *appl, session)
	if err != nil {
		return nil, wrapOktaError(err, core.ErrSAMLError)
	}

	return samlResponse, nil
}

var _ core.AuthenticationProvider = &Authenticator{}

// New creates a new Okta authenticator.
// An error may be returned if the token or host are in the incorrect format - please refer to the Okta documentation at github.com/okta/okta-sdk-golang/
func New(host, token string, mfa duo.Duo) (*Authenticator, error) {
	// This is a bit of a hack, but if we assume that the URL passed will always a hostname, we must add https:// ourselves
	// If we do not add https, Okta will complain, and if we do, our own code will break
	uri := url.URL{Host: host, Scheme: "https"}
	ctx, client, err := okta.NewClient(context.Background(), okta.WithOrgUrl(uri.String()), okta.WithToken(token))
	if err != nil {
		return nil, err
	}

	return &Authenticator{client: client, ctx: ctx, mfa: mfa, oktaAuthClient: newOktaAuthClient(host)}, nil
}

func Must(host, token string, mfa duo.Duo) *Authenticator {
	auth, err := New(host, token, mfa)
	if err != nil {
		panic(err)
	}

	return auth
}

// translateOktaError converts an error from Okta to one of the standard provider's errors.
// If the function can't translate the error, it returns a specified default error.
func translateOktaError(err error, defaultErr core.AuthenticationProviderError) core.AuthenticationProviderError {
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
func wrapOktaError(err error, defaultCoreErr core.AuthenticationProviderError) error {
	return core.WrapError(translateOktaError(err, defaultCoreErr), err)
}
