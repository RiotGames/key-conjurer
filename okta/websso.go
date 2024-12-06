package okta

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/RobotsAndPencils/go-saml"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"
)

var ErrNoSAMLAssertion = errors.New("no saml assertion")

// exchangeAccessTokenForWebSSOToken exchanges an OAuth2 token for an Okta Web SSO token.
//
// An Okta Web SSO token is a non-standard authorization token for Okta's Web SSO endpoint.
func exchangeAccessTokenForWebSSOToken(ctx context.Context, oauthCfg *oauth2.Config, accessToken string, idToken string, applicationID string) (*oauth2.Token, error) {
	return oauthCfg.Exchange(ctx, "",
		oauth2.SetAuthURLParam("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange"),
		oauth2.SetAuthURLParam("actor_token", accessToken),
		oauth2.SetAuthURLParam("actor_token_type", "urn:ietf:params:oauth:token-type:access_token"),
		oauth2.SetAuthURLParam("subject_token", idToken),
		oauth2.SetAuthURLParam("subject_token_type", "urn:ietf:params:oauth:token-type:id_token"),
		// https://www.linkedin.com/pulse/oktas-aws-cli-app-mysterious-case-powerful-okta-apis-chaim-sanders/
		oauth2.SetAuthURLParam("requested_token_type", "urn:okta:oauth:token-type:web_sso_token"),
		oauth2.SetAuthURLParam("audience", fmt.Sprintf("urn:okta:apps:%s", applicationID)),
	)
}

// exchangeWebSSOTokenForSAMLAssertion is an Okta-specific API which exchanges an Okta Web SSO token, which is obtained by exchanging an OAuth2 token using the RFC8693 Token Exchange Flow, for a SAML assertion.
//
// It is not standards compliant, but is used by Okta in their own okta-aws-cli.
func exchangeWebSSOTokenForSAMLAssertion(ctx context.Context, issuer string, token *oauth2.Token) ([]byte, error) {
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

func ExchangeTokenForAssertion(ctx context.Context, cfg *oauth2.Config, accessToken, idToken, oidcDomain, applicationID string) (*saml.Response, string, error) {
	tok, err := exchangeAccessTokenForWebSSOToken(ctx, cfg, accessToken, idToken, applicationID)
	if err != nil {
		return nil, "", fmt.Errorf("error exchanging token: %w", err)
	}

	assertionBytes, err := exchangeWebSSOTokenForSAMLAssertion(ctx, oidcDomain, tok)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch SAML assertion: %w", err)
	}

	response, err := saml.ParseEncodedResponse(string(assertionBytes))
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse SAML response: %w", err)
	}

	return response, string(assertionBytes), nil
}
