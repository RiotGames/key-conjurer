package okta

import (
	"context"
	"errors"
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
func (a *Authenticator) Authenticate(ctx context.Context, creds core.Credentials) (core.User, error) {
	req := authnRequest{Username: creds.Username, Password: creds.Password}
	res, err := a.oktaAuthClient.Authn(ctx, req)
	// We don't need to acknowledge this error because we're using zero values all the way down
	return core.User{ID: res.UserID()}, err
}

// ListApplications should list all the applications the given user is entitled to access.
func (a *Authenticator) ListApplications(ctx context.Context, user core.User) ([]core.Application, error) {
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

	var awsAccounts []core.Application
	for _, app := range links {
		if app.AppName != "amazon_aws" {
			continue
		}

		awsAccounts = append(awsAccounts, core.Application{
			LegacyID: 0,
			ID:       app.AppInstanceId,
			Name:     app.Label,
		})
	}

	return awsAccounts, nil
}

type extractedRole struct {
	AWSAccountName string
	AWSAccountID   string
	AWSRoleName    string
	OktaGroupID    string
}

// extractAWSAccountName attempts to extract the account name and ID from the given Okta group.
func extractRole(group *okta.Group) (extractedRole, bool) {
	// TODO: Filtering the format this way in this location seems like a bad idea.
	// It's certainly not trivial for other people to use.
	var r extractedRole

	// RG-AWS.account_name.role_name.account_id
	split := strings.Split(group.Profile.Name, ".")
	if split[0] != "RG-AWS" || len(split) != 4 {
		return r, false
	}

	r.AWSAccountName = split[1]
	r.AWSRoleName = split[2]
	r.AWSAccountID = split[3]
	r.OktaGroupID = group.Id
	return r, true
}

// GenerateSAMLAssertion should generate a SAML assertion that the user may exchange with the target application in order to gain access to it.
// This will initiate a multi-factor request with Duo.
func (a *Authenticator) GenerateSAMLAssertion(ctx context.Context, creds core.Credentials, appID string) (*core.SAMLResponse, error) {
	if appID == "" {
		return nil, errors.New("appID cannot be an empty string")
	}

	app, _, err := a.client.Application.GetApplication(ctx, appID, &okta.Application{}, query.NewQueryParams())
	if err != nil {
		return nil, err
	}

	appl := app.(*okta.Application)

	st, err := a.oktaAuthClient.Authn(ctx, authnRequest{Username: creds.Username, Password: creds.Password})
	if err != nil {
		return nil, err
	}

	var f *okta.UserFactor
	for _, factor := range st.Factors() {
		if factor.Provider == "DUO" && factor.FactorType == "web" {
			f = &factor
			break
		}
	}

	if f == nil {
		return nil, errors.New("no Duo web factor found")
	}

	vf, err := a.oktaAuthClient.VerifyFactor(ctx, st.StateToken, *f)
	if err != nil {
		return nil, err
	}

	tok, err := a.mfa.SendPush(vf.AuthSignature, vf.StateToken.String(), vf.CallbackURL, vf.Host)
	if err != nil {
		return nil, err
	}

	if err = a.oktaAuthClient.SubmitVerifyFactorResponse(ctx, vf, tok); err != nil {
		return nil, err
	}

	session, err := a.oktaAuthClient.CreateSession(ctx, vf)
	if err != nil {
		return nil, err
	}

	return a.oktaAuthClient.GetSAMLResponse(ctx, *appl, session)
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
