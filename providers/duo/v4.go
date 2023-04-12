package duo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/riotgames/key-conjurer/pkg/htmlutil"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
)

func executeAndParseJSON[T any](client *http.Client, req *http.Request) (dst T, err error) {
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&dst)
	return
}

// DuoV4 contains functions for interacting with the Duo v4 Frame API.
type DuoV4 struct {
	BaseURL *url.URL
	Client  *http.Client
}

func (d DuoV4) CheckStatusRequest(ctx context.Context, session Session) (*http.Request, error) {
	vals := url.Values{
		"txid": []string{session.Txid},
		"sid":  []string{session.Sid},
	}

	next := d.BaseURL.ResolveReference(&url.URL{Path: "/frame/v4/status"})
	req, err := http.NewRequestWithContext(ctx, "POST", next.String(), strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

type CheckStatusResponse struct{}

// CheckStatus checks the MFA status of the given txid/sid in Duo.
//
// The return type is not implemented because it is discarded by KeyConjurer.
func (d DuoV4) CheckStatus(ctx context.Context, session Session) (CheckStatusResponse, error) {
	req, _ := d.CheckStatusRequest(ctx, session)
	return executeAndParseJSON[CheckStatusResponse](d.Client, req)
}

func (d DuoV4) WaitForPushAcknowledgement(ctx context.Context, session Session) error {
	// TODO: Handle rejections - If we don't handle rejections, somewhere later in the application will fail with a cryptic error.
	// This may not require two check statuses, if one was already called - but in normal functioning, this will always require at least two.
	_, err := d.CheckStatus(ctx, session)
	if err != nil {
		return err
	}

	// The second one blocks until the push is approved.
	// This second request will include cookies that we need to use in subsequent requests.
	_, err = d.CheckStatus(ctx, session)
	return err
}

type DuoV4APIResponse[T any] struct {
	Stat     string
	Response T
}

type Factor struct {
	Key   string
	Index string
}

type RegisteredFactors struct {
	Phones []Factor
}

type Session struct {
	Sid  string
	Txid string
	Xsrf string
}

// AuthFromResponse is like Open, but it consumes the given response rather than issuing one.
//
// This is useful when you know you have been redirected to the Duo session Auth page.
func (d DuoV4) AuthFromResponse(ctx context.Context, resp *http.Response) (Session, error) {
	// This is the /frame/frameless/v4/auth stage.
	body, err := html.Parse(resp.Body)
	if err != nil {
		return Session{}, err
	}

	form, ok := htmlutil.FindFirstForm(body)
	if !ok {
		// Couldn't find a form - This is probably the wrong web page, or the format changed somehow
		return Session{}, ErrCouldNotFindDuoForm
	}

	// Submit initial auth request
	req, _ := http.NewRequestWithContext(ctx, "POST", resp.Request.URL.String(), strings.NewReader(form.Values().Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")

	resp, err = d.Client.Do(req)
	if err != nil {
		return Session{}, err
	}

	body, err = html.Parse(resp.Body)
	if err != nil {
		return Session{}, err
	}

	form, ok = htmlutil.FindFirstForm(body)
	if !ok {
		return Session{}, ErrCouldNotFindDuoForm
	}

	// Submit client certificates
	req, _ = http.NewRequestWithContext(ctx, "POST", resp.Request.URL.String(), strings.NewReader(form.Values().Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")

	resp, err = d.Client.Do(req)
	if err != nil {
		return Session{}, err
	}

	body, err = html.Parse(resp.Body)
	if err != nil {
		return Session{}, err
	}

	form, ok = htmlutil.FindFirstForm(body)
	if !ok {
		return Session{}, ErrCouldNotFindDuoForm
	}

	session := Session{
		Sid:  form.Inputs["sid"],
		Xsrf: form.Inputs["_xsrf"],
	}

	return session, nil
}

func (d DuoV4) GetRegisteredFactorDataRequest(ctx context.Context, session Session) (*http.Request, error) {
	vals := url.Values{
		"sid":              []string{session.Sid},
		"post_auth_action": []string{"OIDC_EXIT"},
	}

	next := d.BaseURL.ResolveReference(&url.URL{Path: "/frame/v4/auth/prompt/data"})
	next.RawQuery = vals.Encode()
	return http.NewRequestWithContext(ctx, "GET", next.String(), nil)
}

func (d DuoV4) GetRegisteredFactors(ctx context.Context, session Session) ([]Factor, error) {
	dataReq, _ := d.GetRegisteredFactorDataRequest(ctx, session)
	resp, err := d.Client.Do(dataReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data DuoV4APIResponse[RegisteredFactors]
	return data.Response.Phones, json.NewDecoder(resp.Body).Decode(&data)
}

// Exit ends the current OIDC session. The returned http.Response will contain the response of whereever Duo redirected us, usually to the Redirect URI of the OIDC application.
func (d DuoV4) Exit(ctx context.Context, session Session, deviceKey string) (*http.Response, error) {
	v := url.Values{
		"sid":           []string{session.Sid},
		"txid":          []string{session.Txid},
		"_xsrf":         []string{session.Xsrf},
		"device_key":    []string{deviceKey},
		"factor":        []string{"Duo Push"},
		"dampen_choice": []string{"false"},
	}

	next := d.BaseURL.ResolveReference(&url.URL{Path: "/frame/v4/oidc/exit"})
	req, _ := http.NewRequestWithContext(ctx, "POST", next.String(), strings.NewReader(v.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return d.Client.Do(req)
}

// Prompt issues a prompt for the given session to the given factor and device.
//
// A new Session is returned, which must be used for future calls.
func (d DuoV4) Prompt(ctx context.Context, session Session, factor, device string) (Session, error) {
	vals := url.Values{
		"sid":    []string{session.Sid},
		"_xsrf":  []string{session.Xsrf},
		"factor": []string{factor},
		"device": []string{device},
	}

	next := d.BaseURL.ResolveReference(&url.URL{Path: "/frame/v4/prompt"})
	req, _ := http.NewRequestWithContext(ctx, "POST", next.String(), strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := d.Client.Do(req)
	if err != nil {
		return Session{}, err
	}

	defer resp.Body.Close()
	blob, err := io.ReadAll(resp.Body)
	if err != nil {
		return Session{}, err
	}

	session.Txid = gjson.GetBytes(blob, "response.txid").Str
	return session, nil
}

func (d DuoV4) PromptPhone1(ctx context.Context, session Session) (Session, error) {
	return d.Prompt(ctx, session, "Duo Push", "phone1")
}
