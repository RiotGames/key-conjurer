package okta

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/riotgames/key-conjurer/api/core"
	"github.com/valyala/fastjson"
)

// oktaAuthClient encapsulates authentication operations that are not exposed by the regular Okta client
type oktaAuthClient struct {
	url url.URL
	rt  http.RoundTripper
}

type jsonHAL map[string]json.RawMessage

func (j jsonHAL) Get(key string, dest interface{}) error {
	return json.Unmarshal(j[key], dest)
}

type authnRequest struct {
	Audience string `json:"audience,omitempty"`
	Username string `json:"username"`
	Password string `json:"password"`
	Options  struct {
		MultiOptionalFactorEnroll bool `json:"multiOptionalFactorEnroll"`
		WarnBeforePasswordExpired bool `json:"warnBeforePasswordExpired"`
	} `json:"options"`
}

type sessionToken string

func (s *sessionToken) String() string { return string(*s) }

type stateToken string

func (s *stateToken) String() string { return string(*s) }

type authnResponse struct {
	Status     string
	ExpiresAt  time.Time
	StateToken stateToken
	Embedded   jsonHAL `json:"_embedded"`
	Links      jsonHAL `json:"_links"`
}

func (p *authnResponse) Factors() []okta.UserFactor {
	var factors []okta.UserFactor
	p.Embedded.Get("factors", &factors)
	return factors
}

func (p *authnResponse) UserID() string {
	var user struct {
		ID string `json:"id"`
	}

	// Error intentionally ignored.
	p.Embedded.Get("user", &user)
	return user.ID
}

func newOktaAuthClient(hostname string) oktaAuthClient {
	var baseOktaURL = url.URL{
		Scheme: "https",
		Host:   hostname,
	}

	return oktaAuthClient{url: baseOktaURL, rt: http.DefaultTransport}
}

func (o *oktaAuthClient) do(ctx context.Context, method, path string, data, result interface{}) (response *http.Response, err error) {

	b, err := json.Marshal(data)
	if err != nil {
		err = wrapError(err, ErrOktaBadRequest)
		return
	}

	uri := o.url
	uri.Path = path
	req, err := http.NewRequestWithContext(ctx, method, uri.String(), bytes.NewReader(b))
	if err != nil {
		err = wrapError(err, ErrOktaBadRequest)
		return
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")

	response, err = o.rt.RoundTrip(req)
	if err != nil {
		return
	}

	dec := json.NewDecoder(response.Body)
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		err = dec.Decode(result)
		return
	default:
		var inner okta.Error
		err = dec.Decode(&inner)
		if err != nil {
			return
		}

		err = &inner
		return
	}
}

// Authn posts a request to the /authn endpoint.
func (o *oktaAuthClient) Authn(ctx context.Context, payload authnRequest) (authnResponse, error) {
	var res authnResponse
	httpResponse, err := o.do(ctx, "POST", "/api/v1/authn", payload, &res)
	return res, translateError(httpResponse, err)
}

type verifyFactorResponse struct {
	CallbackURL, CancelURL, Status, Host string
	StateToken                           stateToken
	Deadline                             time.Time
	AuthSignature, AppSignature          string
	DeviceID                             string
}

// ParseJSON attempts to parse response from the given parser.
func (v *verifyFactorResponse) ParseJSON(b []byte, p fastjson.Parser) error {
	// This function will convert []byte arrays to strings.
	// This ensures that they will stick around in memory after the next fastjson.Parser call.
	val, err := p.ParseBytes(b)
	if err != nil {
		return err
	}

	if err := v.Deadline.UnmarshalText(val.GetStringBytes("expiresAt")); err != nil {
		return err
	}

	v.DeviceID = string(val.GetStringBytes("_embedded", "factor", "id"))
	v.Status = string(val.GetStringBytes("status"))
	v.StateToken = stateToken(val.GetStringBytes("stateToken"))
	v.CancelURL = string(val.GetStringBytes("_embedded", "_links", "cancel", "href"))

	verif := val.Get("_embedded", "factor", "_embedded", "verification")
	v.Host = string(verif.GetStringBytes("host"))
	v.CallbackURL = string(verif.GetStringBytes("_links", "complete", "href"))
	signature := strings.Split(string(verif.GetStringBytes("signature")), ":")
	v.AuthSignature = signature[0]
	v.AppSignature = signature[1]
	return nil
}

func (o *oktaAuthClient) VerifyFactor(ctx context.Context, token stateToken, factor okta.UserFactor) (verifyFactorResponse, error) {
	var p fastjson.Parser
	var data verifyFactorResponse
	var r json.RawMessage

	type t struct {
		StateToken string `json:"stateToken"`
	}

	path := fmt.Sprintf("/api/v1/authn/factors/%s/verify", factor.Id)
	httpResponse, err := o.do(ctx, "POST", path, t{StateToken: token.String()}, &r)
	if err != nil {
		return data, translateError(httpResponse, err)
	}

	return data, translateError(nil, data.ParseJSON(r, p))
}

func (o *oktaAuthClient) SubmitChallengeResponse(ctx context.Context, vf verifyFactorResponse, token string) error {
	uri, err := url.Parse(vf.CallbackURL)
	if err != nil {
		return wrapError(fmt.Errorf("unable to parse Callback URL: %s", err), ErrOktaBadRequest)
	}

	v := url.Values{}
	v.Set("id", vf.DeviceID)
	v.Set("stateToken", vf.StateToken.String())
	v.Set("sig_response", fmt.Sprintf("%s:%s", token, vf.AppSignature))
	req, err := http.NewRequestWithContext(ctx, "POST", uri.String(), strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.AddCookie(&http.Cookie{Name: "oktaStateToken", Value: vf.StateToken.String()})
	resp, err := o.rt.RoundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	default:
		var e okta.Error
		if err := dec.Decode(&e); err != nil {
			return err
		}

		return &e
	}
}

type session struct {
	Status       string
	ExpiresAt    time.Time
	SessionToken sessionToken `json:"sessionToken"`
}

func (o *oktaAuthClient) CreateSession(ctx context.Context, vf verifyFactorResponse) (session, error) {
	type t struct {
		StateToken string `json:"stateToken"`
	}

	var s session
	path := fmt.Sprintf("/api/v1/authn/factors/%s/verify", vf.DeviceID)
	httpResponse, err := o.do(ctx, "POST", path, t{StateToken: vf.StateToken.String()}, &s)
	if err != nil {
		return s, translateError(httpResponse, err)
	}

	if s.Status != "SUCCESS" {
		return s, wrapError(fmt.Errorf("could not create session - okta indicates %s", s.Status), ErrOktaCouldNotCreateSession)
	}

	return s, nil
}

func getHrefLink(app okta.Application) (string, bool) {
	links := app.Links.(map[string]interface{})
	appLinks := links["appLinks"].([]interface{})
	for _, interf := range appLinks {
		entry := interf.(map[string]interface{})
		if entry["type"] == "text/html" {
			return entry["href"].(string), true
		}
	}

	return "", false
}

// GetSAMLResponse attempts to initiate a SAML request for the given Application.
func (o *oktaAuthClient) GetSAMLResponse(ctx context.Context, application okta.Application, session session) (*core.SAMLResponse, error) {
	endpoint, ok := getHrefLink(application)
	if !ok {
		return nil, wrapError(errors.New("no endpoint found - this usually indicates an error in the Okta API or Okta SDK"), ErrOktaSAMLError)
	}

	uri, err := url.Parse(endpoint)
	if err != nil {
		return nil, wrapError(err, ErrOktaSAMLError)
	}

	uri.RawQuery = url.Values{"sessionToken": []string{session.SessionToken.String()}}.Encode()

	// We use a custom client to ensure that we do not follow redirects, as this will break the flow.
	client := http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			// Indicate to Go that it must not follow redirects.
			return http.ErrUseLastResponse
		},
		Transport: o.rt,
	}

	// This request will give us a session cookie that we can use.
	resp, err := client.Get(uri.String())
	if err != nil {
		return nil, translateError(resp, err)
	}

	if resp.StatusCode != http.StatusFound {
		return nil, wrapError(fmt.Errorf("okta returned a status code of %d for endpoint %s instead of %d", resp.StatusCode, uri.Path, http.StatusFound), ErrOktaSAMLError)
	}

	req, err := http.NewRequest("GET", resp.Header.Get("Location"), nil)
	if err != nil {
		return nil, wrapError(err, ErrOktaSAMLError)
	}

	var sid *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "sid" {
			sid = cookie
			break
		}
	}

	req.AddCookie(sid)
	// This response will redirect us to the SAML Endpoint
	resp, err = client.Do(req)
	if err != nil {
		return nil, translateError(resp, err)
	}

	// Now we have the SAML URL, we can send a request to that URL with our cookie to get the SAML response
	req, err = http.NewRequest("GET", resp.Header.Get("Location"), nil)
	if err != nil {
		return nil, wrapError(err, ErrOktaSAMLError)
	}

	req.AddCookie(sid)
	resp, err = client.Do(req)
	if err != nil {
		return nil, translateError(resp, err)
	}

	samlResponse, err := extractSAMLResponse(resp.Body)
	if err != nil {
		return samlResponse, wrapError(err, ErrOktaSAMLError)
	}

	return samlResponse, nil
}

// OktaError is an error from Okta.
type OktaError error

// A list of standard errors that can be returned by by Okta client.
var (
	ErrOktaBadRequest            OktaError = errors.New("bad request")
	ErrOktaUnauthorized          OktaError = errors.New("unauthorized")
	ErrOktaForbidden             OktaError = errors.New("forbidden")
	ErrOktaCouldNotCreateSession OktaError = errors.New("could not create a session")
	ErrOktaSAMLError             OktaError = errors.New("could not get a SAML response")
	ErrOktaInternalServerError   OktaError = errors.New("internal server error")
	ErrOktaUnspecified           OktaError = errors.New("unspecified")
)

// wrapError wraps an error into a standard Okta client error.
func wrapError(err error, oktaErr OktaError) error {
	return fmt.Errorf("%w: %s", oktaErr, err.Error())
}

// getMessage extracts a message from an error.
// If it is an error from Okta, the function returns its summary.
// Otherwise, it returns err.Error().
func getMessage(err error) string {
	var oktaErr *okta.Error
	if errors.As(err, &oktaErr) {
		return oktaErr.ErrorSummary
	}
	return err.Error()
}

// translateError converts a specified error to one of the standard Okta client errors.
// The function does not use error codes from Okta API.
// Instead, it may use a specified HTTP response to determine the best standard error.
func translateError(httpResponse *http.Response, err error) error {
	if err == nil {
		return nil
	}

	var standardErr OktaError
	switch {
	case httpResponse == nil:
		standardErr = ErrOktaInternalServerError
	case httpResponse.StatusCode >= 500:
		standardErr = ErrOktaInternalServerError
	case httpResponse.StatusCode == 400:
		standardErr = ErrOktaBadRequest
	case httpResponse.StatusCode == 401:
		standardErr = ErrOktaUnauthorized
	case httpResponse.StatusCode == 403:
		standardErr = ErrOktaForbidden
	default:
		standardErr = ErrOktaUnspecified
	}

	return fmt.Errorf("%w: %s", standardErr, getMessage(err))
}
