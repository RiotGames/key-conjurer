package okta

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/riotgames/key-conjurer/pkg/htmlutil"
	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
)

// StateToken is a token extracted from the body of an Okta response.
//
// It is used to track a single session.
type StateToken string

type SignedToken struct {
	Tx  string
	App string
}

// ConcatenateAuthSignature applies the auth signature to the token, returning a completed challenge response that can be submitted to Okta.
func (tok SignedToken) ConcatenateAuthSignature(authSignature string) string {
	return strings.Join([]string{authSignature, tok.App}, ":")
}

// ParseSignedToken extracts a transaction and application signature from a signed token from Okta.
//
// A signed token is returned from Okta when initiating MFA with Duo.
func ParseSignedToken(token string) (SignedToken, error) {
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return SignedToken{}, errors.New("wrong number of parts in signed token")
	}

	return SignedToken{Tx: parts[0], App: parts[1]}, nil
}

type MultiFactorUpgradeMethod interface {
	Upgrade(ctx context.Context, client *http.Client) (StateToken, error)
}

type AuthenticatorEnrollment struct {
	Key            string
	ContextualData struct {
		Host        string
		SignedToken string
	}
}

type Remediation struct {
	Name   string
	Type   string
	Href   string
	Method string
}

type IdentifyResponse struct {
	StateHandle                    string
	CurrentAuthenticatorEnrollment struct {
		Type  string
		Value AuthenticatorEnrollment
	}

	Remediation struct {
		Type  string
		Value []Remediation
	}
}

// DetermineUpgradePath determines which multi-factor authentication method the current user should avail of to upgrade their session.
func DetermineUpgradePath(resp IdentifyResponse) (MultiFactorUpgradeMethod, bool) {
	// authMethods is a list of all of the authentication methods available to a user to upgrade their session into a fully authenticated one.
	// This list's order is not guaranteed, so we loop through it twice: Once to find the first available authenticator of type OIDC, and next to find the first available authenticator with the name challenge-authenticator.
	// This ensures that we always prefer OIDC authenticators first (which is the DuoFrameless flow), and only fall back to challenge-authenticator (which is the DuoFrame flow) as a last resort.
	authMethods := resp.Remediation.Value
	for _, rem := range authMethods {
		if rem.Type == "OIDC" {
			return DuoFrameless{remediation: rem}, true
		}
	}

	for _, rem := range authMethods {
		// This is intentionally 'Name' as Type is missing for cases where this is true.
		if rem.Name == "challenge-authenticator" && resp.CurrentAuthenticatorEnrollment.Value.Key == "duo" {
			host := resp.CurrentAuthenticatorEnrollment.Value.ContextualData.Host
			tok, err := ParseSignedToken(resp.CurrentAuthenticatorEnrollment.Value.ContextualData.SignedToken)
			if err != nil {
				break
			}

			return DuoIframe{Host: host, SignedToken: tok, CallbackURL: rem.Href, Method: rem.Method, StateHandle: resp.StateHandle}, true
		}
	}

	return nil, false
}

// ApplicationSAMLSource handles the SP-initiated SAML flow for KeyConjurer.
type ApplicationSAMLSource string

func (source ApplicationSAMLSource) URL() string {
	return string(source)
}

func (source ApplicationSAMLSource) Identify(ctx context.Context, client *http.Client, prevURL *url.URL, username, password string, stateToken StateToken) (identifyResponse IdentifyResponse, err error) {
	identifyVals := map[string]any{
		"credentials": map[string]string{
			"passcode": password,
		},
		"identifier":  username,
		"stateHandle": stateToken,
	}
	identifyBytes, _ := json.Marshal(identifyVals)
	identifyURL := prevURL.ResolveReference(&url.URL{Path: "/idp/idx/identify"})
	resp, err := client.Post(identifyURL.String(), "application/json", bytes.NewReader(identifyBytes))
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusForbidden {
		err = errors.New("user does not have access to this application")
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&identifyResponse)
	return
}

type IntrospectResponse struct {
	Success *struct {
		Name string
		Href string
	}
}

func (source ApplicationSAMLSource) Introspect(ctx context.Context, client *http.Client, prevURL *url.URL, stateToken StateToken) (introspectResponse IntrospectResponse, err error) {
	// We need to run introspect to get our next link.
	// Now use /idp/idx/identify to get the IDP link to follow.
	introspectVals := map[string]any{"stateToken": stateToken}
	introspectBytes, _ := json.Marshal(introspectVals)
	introspectURL := prevURL.ResolveReference(&url.URL{Path: "/idp/idx/introspect"})
	resp, err := client.Post(introspectURL.String(), "application/json", bytes.NewReader(introspectBytes))
	if err != nil {
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&introspectResponse)
	return
}

// findStateToken extracts an Okta state token from the given set of bytes, which must be a HTML response from a page which initiates an Okta session.
//
// Okta uses JavaScript redirections to drive the authentication mechanism within its website. To authenticate with Okta, one requires a state token. KeyConjurer relies on intercepting HTTP response from Okta to capture the SAML response in order to exchange that SAML response for temporary credentials with AWS. Because Riot's Okta configuration does not have redirection URLs set up back to KeyConjurer, we're not able to use a headless browser like Selenium to follow the JavaScript redirects, because the browser would follow redirects to the AWS (or Tencent) console. Therefore, we need to get a state token and execute the flow manually. The only way to retrieve a state token that works in both the legacy and in the OAuth redirection flow is by interrogating a SP-initiated response from Okta, which will contain the state token within an inline script on the page.
//
// This is brittle. If anything within KeyConjurer were to break, it's most likely this portion of the code. We have mentioned our reliance on this idiosyncrasy to Okta, and hopefully they never change it.
func findStateToken(b []byte) (StateToken, bool) {
	matches := stateTokenExpr.FindSubmatch(b)
	idx := stateTokenExpr.SubexpIndex("Token")
	if len(matches) < idx+1 {
		return "", false
	}

	// The JavaScript in the response has HTML-encoded characters in.
	// When a web browser reads it, this is no problem. But, when we read it, it is.
	// Specifically the only character observed so far is \x2D which is a hyphen.
	return StateToken(strings.ReplaceAll(string(matches[idx]), "\\x2D", "-")), true
}

// GetAssertion attempts to retrieve a SAML assertion from the given source and returns the bytes of that assertion.
//
// These bytes are not parsed.
func (source ApplicationSAMLSource) GetAssertion(ctx context.Context, username, password string) ([]byte, error) {
	jar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	client := http.Client{Jar: jar}
	req, _ := http.NewRequest("GET", string(source.URL()), nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send initial application request: %w", err)
	}

	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	stateToken, ok := findStateToken(buf)
	if !ok {
		return nil, fmt.Errorf("could not find state token: %w", err)
	}

	identifyResponse, err := source.Identify(ctx, &client, resp.Request.URL, username, password, stateToken)
	if err != nil {
		return nil, fmt.Errorf("could not call /idp/idx/identify: %w", err)
	}

	method, ok := DetermineUpgradePath(identifyResponse)
	if !ok {
		return nil, errors.New("could determine 2fa upgrade path - users probably does not have a supported device")
	}

	nextStateToken, err := method.Upgrade(ctx, &client)
	if err != nil {
		return nil, fmt.Errorf("could not upgrade session with mfa: %w", err)
	}

	// This type switch violates encapsulation but I've yet to find a "neat" way of encapsulating the Okta-Duo relationship in a single type that isn't a leaky abstraction.
	switch method.(type) {
	case DuoIframe:
		// HACK: In the live version of Okta, a different series of events is followed, where the correct URL obtained by following /idp/idx/introspect.
		// However, we can also issue a request to the application endpoint again, because at this point we have a valid session in our http.Client cookie jar.
		req, _ = http.NewRequest("GET", string(source.URL()), nil)
		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not send initial application request: %w", err)
		}

		defer resp.Body.Close()

		doc, _ := html.Parse(resp.Body)
		appForm, ok := htmlutil.FindFormByID(doc, "appForm")
		if !ok {
			return nil, errors.New("could not find SAMLResponse within response from Okta")
		}

		return []byte(appForm.Inputs["SAMLResponse"]), nil
	case DuoFrameless:
		ixResp, err := source.Introspect(ctx, &client, resp.Request.URL, nextStateToken)
		if err != nil {
			return nil, err
		}

		req, _ = http.NewRequest("GET", ixResp.Success.Href, nil)
		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not fetch SAML response: %w", err)
		}

		defer resp.Body.Close()
		doc, _ := html.Parse(resp.Body)
		appForm, ok := htmlutil.FindFormByID(doc, "appForm")
		if !ok {
			return nil, errors.New("could not find SAMLResponse within response from Okta")
		}
		return []byte(appForm.Inputs["SAMLResponse"]), nil
	default:
		panic("not implemented")
	}
}
