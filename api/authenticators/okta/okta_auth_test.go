package okta

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/stretchr/testify/require"
)

func TestGetMessage(t *testing.T) {
	oktaErr := okta.Error{ErrorCode: "E01", ErrorSummary: "okta error"}
	require.Equal(t, "okta error", getMessage(&oktaErr), "expected only error summary")
	err := errors.New("error description")
	require.Equal(t, "error description", getMessage(err))
}

func TestTranslateError(t *testing.T) {
	require.Nil(t, translateError(nil, nil), "expected nil since error is nil")
	oktaErr := translateError(nil, errors.New("error text"))
	require.True(t, errors.Is(oktaErr, ErrOktaInternalServerError), "expected internal server error")
	require.Contains(t, oktaErr.Error(), ErrOktaInternalServerError.Error(), "expected internal server error text")
	require.Contains(t, oktaErr.Error(), "error text", "expected the original error text")
	oktaErr = translateError(&http.Response{StatusCode: 400}, errors.New("failed"))
	require.True(t, errors.Is(oktaErr, ErrOktaBadRequest), "expected bad request error")
	require.Contains(t, oktaErr.Error(), ErrOktaBadRequest.Error(), "expected bad request error text")
	require.Contains(t, oktaErr.Error(), "failed", "expected the original error text")
}

func TestWrapError(t *testing.T) {
	oktaErr := wrapError(errors.New("access denied"), ErrOktaForbidden)
	require.True(t, errors.Is(oktaErr, ErrOktaForbidden), "expected forbidden error")
	require.Contains(t, oktaErr.Error(), ErrOktaForbidden.Error(), "expected forbidden text")
	require.Contains(t, oktaErr.Error(), "access denied", "expected the original error text")
}

// RoundTripperMock implements RoundTripper interface for testing.
type RoundTripperMock struct {
	response *http.Response
	err      error
}

// RoundTrip returns a pre-defined HTTP response and an error.
func (o RoundTripperMock) RoundTrip(request *http.Request) (*http.Response, error) {
	return o.response, o.err
}

var testOktaUrl url.URL = url.URL{}

// MakeBody wraps a string into an io.ReadCloser.
func MakeBody(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}

func TestAuthnWithOktaErrors(t *testing.T) {
	httpTransportMock := RoundTripperMock{}
	httpTransportMock.response = nil
	httpTransportMock.err = errors.New("connection failed")
	client := oktaAuthClient{url: testOktaUrl, rt: httpTransportMock}
	_, oktaErr := client.Authn(context.TODO(), authnRequest{})
	require.True(t, errors.Is(oktaErr, ErrOktaInternalServerError), "expected internal server error")
	require.Contains(t, oktaErr.Error(), ErrOktaInternalServerError.Error(), "expected internal server error message")
	require.Contains(t, oktaErr.Error(), httpTransportMock.err.Error(), "expected the original error message")

	httpTransportMock = RoundTripperMock{}
	oktaReply := `
	{
		"errorCode": "E0000004",
		"errorSummary": "Authentication failed",
		"errorLink": "E0000004",
		"errorId": "sampleXhLJI0ZNxN0ab8IobVb",
		"errorCauses": []
	}
	`
	httpTransportMock.response = &http.Response{StatusCode: 401, Body: MakeBody(oktaReply)}
	httpTransportMock.err = nil
	client = oktaAuthClient{url: testOktaUrl, rt: httpTransportMock}
	_, oktaErr = client.Authn(context.TODO(), authnRequest{})
	require.True(t, errors.Is(oktaErr, ErrOktaUnauthorized), "expected unauthorized error")
	require.Contains(t, oktaErr.Error(), ErrOktaUnauthorized.Error(), "expected unauthorized error message")
	require.Contains(t, oktaErr.Error(), "Authentication failed", "expected error summary from the Okta reply")

	httpTransportMock = RoundTripperMock{}
	oktaReply = `
	{
		"errorCode": "E0000006",
		"errorSummary": "You do not have permission to perform the requested action",
		"errorLink": "E0000006",
		"errorId": "sampleBmGsRUZa0_Nsv82RoOL",
		"errorCauses": []
	}
	`
	httpTransportMock.response = &http.Response{StatusCode: 403, Body: MakeBody(oktaReply)}
	httpTransportMock.err = nil
	client = oktaAuthClient{url: testOktaUrl, rt: httpTransportMock}
	_, oktaErr = client.Authn(context.TODO(), authnRequest{})
	require.True(t, errors.Is(oktaErr, ErrOktaForbidden), "expected forbidden error")
	require.Contains(t, oktaErr.Error(), ErrOktaForbidden.Error(), "expected forbidden message")
	require.Contains(t, oktaErr.Error(), "You do not have permission to perform the requested action", "expected error summary from the Okta reply")
}
