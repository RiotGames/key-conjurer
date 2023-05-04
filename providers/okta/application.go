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
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
	"golang.org/x/net/publicsuffix"
)

var (
	ErrStateTokenNotFound           = errors.New("could not find state token")
	ErrNoSupportedMultiFactorDevice = errors.New("no supported multi-factor device")
	ErrUnauthorized                 = errors.New("unauthorized")
	ErrInternalServerError          = errors.New("internal server error")
	ErrNoSAMLResponseFound          = errors.New("no SAML response found")
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
	Upgrade(ctx context.Context, client *http.Client) ([]byte, error)
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
	Type   *string
	Href   string
	Method string

	// Value can be multiple types which can only be discerned at runtime.
	// In the tested flows we have, this will be identified as either a string property, or an object.
	Value json.RawMessage
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

func findRemediationByType(rems []Remediation, typ string) (Remediation, bool) {
	for _, rem := range rems {
		if rem.Type != nil && *rem.Type == typ {
			return rem, true
		}
	}

	return Remediation{}, false
}

func findRemediationByName(rems []Remediation, name string) (Remediation, bool) {
	for _, rem := range rems {
		if rem.Name == name {
			return rem, true
		}
	}

	return Remediation{}, false
}

func findAuthenticatorByMethodType(rem Remediation, typ string) (string, bool) {
	// This method is a bit rough because we need to interrogate rem.Value, whose type is undefined until runtime.
	// Generally, it looks a bit like this:
	// [{"name": "authenticator": "type": "object", "options": [{"label": "...", "relatesTo": "..."}, {"name": "stateHandle", ....}]
	if !gjson.ValidBytes(rem.Value) {
		return "", false
	}

	var authenticators []gjson.Result
	for _, auth := range gjson.ParseBytes(rem.Value).Array() {
		if auth.Get("name").Str == "authenticator" {
			authenticators = append(authenticators, auth)
		}
	}

	// Find the first id where methodType == idp.
	for _, auth := range authenticators {
		for _, option := range auth.Get("options").Array() {
			formValues := option.Get("value.form.value")
			maybeId := formValues.Get(`#(name=="id").value`)
			maybeMethodType := formValues.Get(`#(name=="methodType").value`)
			if maybeMethodType.Str == typ && maybeId.Exists() {
				return maybeId.Str, true
			}
		}
	}

	return "", false
}

// findFirstSuitableAuthenticatorID returns the first suitable authenticator ID for a user within the given remediation.
//
// This is useful for when a user has multiple  potentially valid authenticators and Okta has indicated that we must pick one.
func findFirstSuitableAuthenticatorID(rem Remediation) (string, bool) {
	id, ok := findAuthenticatorByMethodType(rem, "idp")
	if ok {
		return id, true
	}

	return findAuthenticatorByMethodType(rem, "duo")
}

// errNeedsAuthenticatorSelection indicates that Okta wanted the user to select an authenticator.
//
// This type should not leave the bounds of this package - either we should pick an authenticator for the user, or we should return ErrNoSupportedMultiFactorDevice
type errNeedsAuthenticatorSelection struct {
	Remediation Remediation
}

func (e errNeedsAuthenticatorSelection) Error() string {
	return "user needs to select an authenticator"
}

// DetermineUpgradePath determines which multi-factor authentication method the current user should avail of to upgrade their session.
func DetermineUpgradePath(resp IdentifyResponse, source ApplicationSAMLSource) (MultiFactorUpgradeMethod, error) {
	authMethods := resp.Remediation.Value
	if rem, ok := findRemediationByType(authMethods, "OIDC"); ok {
		return DuoFrameless{remediation: rem, source: source}, nil
	}

	// Legacy Duo Flow
	if rem, ok := findRemediationByName(authMethods, "challenge-authenticator"); ok {
		host := resp.CurrentAuthenticatorEnrollment.Value.ContextualData.Host
		tok, err := ParseSignedToken(resp.CurrentAuthenticatorEnrollment.Value.ContextualData.SignedToken)
		if err != nil {
			// This should never happen
			return nil, fmt.Errorf("failed to parse signed token from legacy duo flow: %s", err)
		}

		return DuoIframe{Host: host, SignedToken: tok, CallbackURL: rem.Href, Method: rem.Method, StateHandle: resp.StateHandle, InitialURL: source.URL()}, nil
	}

	// This will occur if the user has multiple choices to select from and Duo indicates that we must pick one.
	//
	// We can't just return the ID and be done with it - the caller must explicitly issue a request to RespondToChallenge first before continuing.
	if rem, ok := findRemediationByName(authMethods, "select-authenticator-authenticate"); ok {
		return nil, errNeedsAuthenticatorSelection{Remediation: rem}
	}

	return nil, ErrNoSupportedMultiFactorDevice
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
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		err = ErrUnauthorized
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
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&introspectResponse)
	return
}

var stateTokenExpr = regexp.MustCompile("var stateToken = '(?P<Token>.*)';")

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
		return nil, ErrStateTokenNotFound
	}

	identifyResponse, err := source.Identify(ctx, &client, resp.Request.URL, username, password, stateToken)
	if err != nil {
		return nil, fmt.Errorf("could not call /idp/idx/identify: %w", err)
	}

	method, err := DetermineUpgradePath(identifyResponse, source)
	// If a user has more than one device, Okta may require that the user pick one.
	// If this is the case, an errNeedsAuthenticatorSelection will be returned which we must handle here.
	var errAuthSelection errNeedsAuthenticatorSelection
	if errors.As(err, &errAuthSelection) {
		authenticatorID, ok := findFirstSuitableAuthenticatorID(errAuthSelection.Remediation)
		// We couldn't find a suitable authenticator and the user doesn't have anything else we can use.
		if !ok {
			return nil, ErrNoSupportedMultiFactorDevice
		}

		challResponse, err := source.RespondToChallenge(ctx, &client, resp.Request.URL, authenticatorID, identifyResponse.StateHandle)
		if err != nil {
			return nil, fmt.Errorf("could not call /idp/idx/challenge: %w", err)
		}

		// If the challenge response was a success, we again try to determine the upgrade path based on the returned values.
		method, err = DetermineUpgradePath(challResponse, source)
		if err != nil {
			return nil, ErrNoSupportedMultiFactorDevice
		}
	} else if err != nil {
		return nil, err
	}

	saml, err := method.Upgrade(ctx, &client)
	if err != nil {
		return nil, fmt.Errorf("could not upgrade session with mfa: %w", err)
	}

	return saml, nil
}

func (source ApplicationSAMLSource) RespondToChallenge(ctx context.Context, client *http.Client, prevURL *url.URL, authenticatorID, stateHandle string) (challengeResp IdentifyResponse, err error) {
	vals := map[string]any{"stateHandle": stateHandle, "authenticator": map[string]string{"id": authenticatorID}}
	buf, _ := json.Marshal(vals)
	uri := prevURL.ResolveReference(&url.URL{Path: "/idp/idx/challenge"})
	resp, err := client.Post(uri.String(), "application/json", bytes.NewReader(buf))
	if err != nil {
		return
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&challengeResp)
	return
}
