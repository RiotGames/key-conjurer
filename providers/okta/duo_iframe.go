package okta

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/riotgames/key-conjurer/providers/duo"
	"github.com/tidwall/gjson"
)

type DuoIframe struct {
	Host        string
	SignedToken SignedToken
	CallbackURL string
	StateHandle string
	Method      string
	StateToken  StateToken
}

func (f DuoIframe) Upgrade(ctx context.Context, client *http.Client) (StateToken, error) {
	duo := duo.New()
	tok, err := duo.SendPush(f.SignedToken.Tx, string(f.StateToken), f.CallbackURL, f.Host)
	if err != nil {
		return f.StateToken, err
	}

	// Post the state token back to the callbackURL.
	vals := map[string]any{
		"credentials": map[string]string{
			"signatureData": f.SignedToken.ConcatenateAuthSignature(tok),
		},
		"stateHandle": f.StateHandle,
	}

	// HACK: This sequence of events doesn't track with what is experienced in the web browser.
	// In the web browser, vals is sent to f.CallbackURL, like here, but the response is inspected to retrieve the `success.href` property within the body -
	// this is a redirect link which, when followed, should redirect us to the SAML page.
	//
	// However, this appears to always return HTTP 500 and take quite a bit of time to execute.
	//
	// As mentioned in application.go, we can work around this by visiting the application link again, which the caller will do - because our cookie jar in
	// the http.Client is tracking all of our cookies, we have a valid session by the time this call ends, with or without following the redirection, because
	// the session is upgraded with the following request.
	reqBuf, _ := json.Marshal(vals)
	req, _ := http.NewRequestWithContext(ctx, f.Method, f.CallbackURL, bytes.NewReader(reqBuf))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	// We need to get the redirect URI. This redirect will fail with a HTTP 500, but is necessary to ensure the state token is upgraded.
	// The redirect URI can be found at success.href.
	if err != nil {
		return f.StateToken, err
	}

	defer resp.Body.Close()
	bodyBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return f.StateToken, err
	}

	// This call WILL http 500 and that is ok! we just need to ensure it goes through.
	// It's unclear why it HTTP 500s, but the important part is that we issue the request to success.href (which will usually be /redirect with a state token).
	// Okta still sets the appropriate response cookie, allowing us to authenticate.
	req, _ = http.NewRequestWithContext(ctx, "GET", gjson.GetBytes(bodyBuf, "success.href").Str, nil)
	_, err = client.Do(req)
	return f.StateToken, err
}
