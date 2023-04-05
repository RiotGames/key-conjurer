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

func DetermineUpgradePath(resp IdentifyResponse) (MultiFactorUpgradeMethod, bool) {
	// Two iterations is fine, this is a small list
	// Prefer Duo OIDC flow
	for _, rem := range resp.Remediation.Value {
		if rem.Type == "OIDC" {
			return DuoFrameless{remediation: rem}, true
		}
	}

	for _, rem := range resp.Remediation.Value {
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

func extractStateToken(r io.Reader) (StateToken, error) {
	// Extremely cursed
	bodyBuf, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	// Extremely cursed
	matches := stateTokenExpr.FindSubmatch(bodyBuf)
	idx := stateTokenExpr.SubexpIndex("Token")
	if len(matches) < idx+1 {
		return "", errors.New("no match found")
	}

	// The JavaScript in the response has HTML-encoded characters in.
	// When a web browser reads it, this is no problem. But, when we read it, it is.
	// Specifically the only character observed so far is \x2D which is a hyphen.
	return StateToken(strings.ReplaceAll(string(matches[idx]), "\\x2D", "-")), nil
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

	stateToken, err := extractStateToken(resp.Body)
	if err != nil {
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
		appForm, ok := findFormByID(doc, "appForm")
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
		appForm, ok := findFormByID(doc, "appForm")
		if !ok {
			return nil, errors.New("could not find SAMLResponse within response from Okta")
		}
		return []byte(appForm.Inputs["SAMLResponse"]), nil
	default:
		panic("not implemented")
	}
}
