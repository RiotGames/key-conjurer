package okta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/riotgames/key-conjurer/pkg/htmlutil"
	"github.com/riotgames/key-conjurer/providers/duo"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
)

type DuoIframe struct {
	Host        string
	SignedToken SignedToken
	CallbackURL string
	StateHandle string
	Method      string
	StateToken  StateToken
	InitialURL  url.URL
}

func (f DuoIframe) Upgrade(ctx context.Context, client *http.Client) ([]byte, error) {
	duo := duo.New()
	tok, err := duo.SendPush(f.SignedToken.Tx, string(f.StateToken), f.CallbackURL, f.Host)
	if err != nil {
		return nil, err
	}

	// Post the state token back to the callbackURL.
	vals := map[string]any{
		"credentials": map[string]string{
			"signatureData": f.SignedToken.ConcatenateAuthSignature(tok),
		},
		"stateHandle": f.StateHandle,
	}

	// The following request will redirect us back to Okta when we receive a successful response.
	reqBuf, _ := json.Marshal(vals)
	req, _ := http.NewRequestWithContext(ctx, f.Method, f.CallbackURL, bytes.NewReader(reqBuf))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	req, _ = http.NewRequestWithContext(ctx, "GET", gjson.GetBytes(bodyBuf, "success.href").Str, nil)
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		// HACK: The previous HTTP response to success.href may return HTTP 500.
		//
		// The SP-initiated flow using an application link may yield a HTTP 500 on the last redirect.
		// This has been raised with Okta and it's not clear why this occurs.
		// Luckily, the server still upgrades the users session, so we are able to retrieve the SAML response from this response instead.
		req, _ = http.NewRequest("GET", f.InitialURL.String(), nil)
		resp, err = client.Do(req)
	}

	if resp.StatusCode != http.StatusOK {
		// Something went wrong.
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	doc, _ := html.Parse(resp.Body)
	form, ok := htmlutil.FindFormByID(doc, "appForm")
	if !ok {
		return nil, ErrNoSAMLResponseFound
	}

	return []byte(form.Inputs["SAMLResponse"]), nil
}
