package okta

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/riotgames/key-conjurer/pkg/htmlutil"
	"github.com/riotgames/key-conjurer/providers/duo"
	"golang.org/x/net/html"
)

type DuoFrameless struct {
	remediation Remediation
	source      ApplicationSAMLSource
}

func (f DuoFrameless) Upgrade(ctx context.Context, client *http.Client) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, f.remediation.Method, f.remediation.Href, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	d := duo.DuoV4{
		Client:  client,
		BaseURL: resp.Request.URL,
	}

	// The frameless flow does include a new state token in the response.
	resp, err = f.handleFlow(ctx, d, resp)
	if err != nil {
		return nil, err
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

	// There is a bug where SP-initiated flows are sometimes hitting HTTP 500 in the DuoFrameless flow. This appears to have come out of nowhere.
	// Much like in the DuoIframe flow, initiating a request to the original source URL after an otherwise successful login (as indicated by Introspect) will allow us to continue to log in.
	var ixResp IntrospectResponse
	ixResp, err = f.source.Introspect(ctx, client, resp.Request.URL, stateToken)
	if err != nil {
		return nil, err
	}

	req, _ = http.NewRequest("GET", ixResp.Success.Href, nil)
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// If this http 500s, we need to issue a second request to the original source URL.
	if resp.StatusCode == http.StatusInternalServerError {
		req, _ = http.NewRequest("GET", string(f.source.URL()), nil)
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusInternalServerError {
			return nil, ErrInternalServerError
		}
	}

	doc, _ := html.Parse(resp.Body)
	form, ok := htmlutil.FindFormByID(doc, "appForm")
	if !ok {
		return nil, ErrNoSAMLResponseFound
	}

	return []byte(form.Inputs["SAMLResponse"]), nil
}

// handleFramelessDuoFlow handles a Duo-type remediation flow from Okta using the OIDC duo flow.
//
// The http.Response for the final call which redirected us back to Okta will be returned.
func (f DuoFrameless) handleFlow(ctx context.Context, d duo.DuoV4, resp *http.Response) (*http.Response, error) {
	session, err := d.AuthFromResponse(ctx, resp)
	if err != nil {
		return nil, err
	}

	session, err = d.PromptPhone1(ctx, session)
	if err != nil {
		return nil, err
	}

	if err := d.WaitForPushAcknowledgement(ctx, session); err != nil {
		return nil, err
	}

	// If the status was approved, we make a request to /exit, which will redirect us.
	// This needs a few extra values - one of which is the device key for phone1.
	// We get the device key from a request to a data endpoint.
	factors, err := d.GetRegisteredFactors(ctx, session)
	if err != nil {
		return nil, err
	}

	// TODO: The correct one can be found by searching through the array for index=="phone1"
	return d.Exit(ctx, session, factors[0].Key)
}
