package oktawebsso

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
	"golang.org/x/oauth2"
)

// ErrNoSAMLAssertion indicates that the sso endpoint had no SAML assertion in its response.
var (
	ErrNoSAMLAssertion = errors.New("no saml assertion")
	// ErrNotOIDCToken indicates that the token provided was not an OIDC token and thus cannot be used with the RFC8693 exchange flow.
	ErrNotOIDCToken = errors.New("not oidc token")
)

// WebSSOToken is a special type of oauth2 token used in Okta's undocumented "web sso" login flow.
type WebSSOToken *oauth2.Token

// ExchangeAccessToken exchanges an OAuth2 token for an Okta Web SSO token.
//
// The OAuth2 token source must have been retrieved from an OIDC flow, or ErrNotOIDCToken will be returned.
//
// An Okta Web SSO token is a non-standard authorization token for Okta's Web SSO endpoint.
func ExchangeAccessToken(ctx context.Context, cfg *oauth2.Config, ts oauth2.TokenSource, applicationID string) (WebSSOToken, error) {
	at, err := ts.Token()
	if err != nil {
		return nil, err
	}

	idToken, ok := at.Extra("id_token").(string)
	if !ok {
		return nil, ErrNotOIDCToken
	}

	return cfg.Exchange(ctx, "",
		oauth2.SetAuthURLParam("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange"),
		oauth2.SetAuthURLParam("actor_token", at.AccessToken),
		oauth2.SetAuthURLParam("actor_token_type", "urn:ietf:params:oauth:token-type:access_token"),
		oauth2.SetAuthURLParam("subject_token", idToken),
		oauth2.SetAuthURLParam("subject_token_type", "urn:ietf:params:oauth:token-type:id_token"),
		// https://www.linkedin.com/pulse/oktas-aws-cli-app-mysterious-case-powerful-okta-apis-chaim-sanders/
		oauth2.SetAuthURLParam("requested_token_type", "urn:okta:oauth:token-type:web_sso_token"),
		oauth2.SetAuthURLParam("audience", fmt.Sprintf("urn:okta:apps:%s", applicationID)),
	)
}

// GetSAMLAssertion is an Okta-specific API which exchanges an Okta Web SSO token, which is obtained by exchanging an OAuth2 token using the RFC8693 Token Exchange Flow, for a SAML assertion.
//
// It is not standards compliant, but is used by Okta in their own okta-aws-cli.
func GetSAMLAssertion(ctx context.Context, issuer string, token WebSSOToken) ([]byte, error) {
	data := url.Values{"token": {token.AccessToken}}
	uri := fmt.Sprintf("%s/login/token/sso?%s", issuer, data.Encode())
	req, _ := http.NewRequestWithContext(ctx, "GET", uri, nil)
	req.Header.Add("Accept", "text/html")

	client := http.DefaultClient
	if val, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		client = val
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return nil, errors.New("internal okta error occurred")
	}

	doc, _ := html.Parse(resp.Body)
	form, ok := FindFirstForm(doc)
	if !ok {
		return nil, ErrNoSAMLAssertion
	}

	saml, ok := form.Inputs["SAMLResponse"]
	if !ok {
		return nil, ErrNoSAMLAssertion
	}

	return []byte(saml), nil
}
