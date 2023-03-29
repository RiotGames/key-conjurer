package okta

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/okta/okta-sdk-golang/v2/okta"
)

type jsonHAL map[string]json.RawMessage

func (j jsonHAL) Get(key string, dest interface{}) error {
	return json.Unmarshal(j[key], dest)
}

// AuthClient encapsulates authentication operations that are not exposed by the regular Okta client
type AuthClient struct {
	url url.URL
	rt  RequestExecutor
}

type RequestExecutor interface {
	Do(*http.Request) (*http.Response, error)
}

type AuthRequest struct {
	Audience string `json:"audience,omitempty"`
	Username string `json:"username"`
	Password string `json:"password"`
	Options  struct {
		MultiOptionalFactorEnroll bool `json:"multiOptionalFactorEnroll"`
		WarnBeforePasswordExpired bool `json:"warnBeforePasswordExpired"`
	} `json:"options"`
}

type AuthResponse struct {
	Status     string
	ExpiresAt  time.Time
	StateToken StateToken
	Embedded   jsonHAL `json:"_embedded"`
	Links      jsonHAL `json:"_links"`
}

func (p AuthResponse) Factors() []okta.UserFactor {
	var factors []okta.UserFactor
	p.Embedded.Get("factors", &factors)
	return factors
}

func (p AuthResponse) UserID() string {
	var user struct {
		ID string `json:"id"`
	}

	// Error intentionally ignored.
	p.Embedded.Get("user", &user)
	return user.ID
}

func NewAuthClient(hostname string, opts ...func(*AuthClient)) AuthClient {
	client := AuthClient{
		url: url.URL{
			Scheme: "https",
			Host:   hostname,
		},
		rt: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(&client)
	}

	return client
}

func WithHTTPClient(client *http.Client) func(*AuthClient) {
	return func(a *AuthClient) {
		a.rt = client
	}
}

func (o AuthClient) do(ctx context.Context, method, path string, data, result interface{}) (response *http.Response, err error) {
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

	response, err = o.rt.Do(req)
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
func (o AuthClient) VerifyCredentials(ctx context.Context, payload AuthRequest) (AuthResponse, error) {
	var res AuthResponse
	httpResponse, err := o.do(ctx, "POST", "/api/v1/authn", payload, &res)
	return res, translateError(httpResponse, err)
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

// OktaError is an error from Okta.
type OktaError error

// A list of standard errors that can be returned by by Okta client.
var (
	ErrOktaBadRequest                   OktaError = errors.New("bad request")
	ErrOktaUnauthorized                 OktaError = errors.New("unauthorized")
	ErrOktaForbidden                    OktaError = errors.New("forbidden")
	ErrOktaCouldNotStepUpAuthentication OktaError = errors.New("could not create a session")
	ErrOktaSAMLError                    OktaError = errors.New("could not get a SAML response")
	ErrOktaInternalServerError          OktaError = errors.New("internal server error")
	ErrOktaUnspecified                  OktaError = errors.New("unspecified")
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
