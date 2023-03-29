package duo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var httpResponse string

var testServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, httpResponse)
}))
var testServerHostName = testServer.URL[8:]
var testClient = testServer.Client()

func Test_Duo_getSid(t *testing.T) {
	type ExpectedOutput struct {
		sid string
		err error
	}

	assert := assert.New(t)

	tests := []struct {
		name           string
		httpResponse   string
		expectedOutput ExpectedOutput
	}{
		{
			name:           "sid is missing",
			httpResponse:   `<html><input name="not_sid" value="foo" /></html>`,
			expectedOutput: ExpectedOutput{sid: "", err: ErrorCannotFindSid}},
		{
			name:           "sid and value with double quotes ",
			httpResponse:   `<html><input name="sid" value="foo"></html>`,
			expectedOutput: ExpectedOutput{sid: "foo", err: nil}},
		{
			name:           "sid with double quotes, value with single quotes",
			httpResponse:   `<html><input name="sid" value='foo'></html>`,
			expectedOutput: ExpectedOutput{sid: "foo", err: nil}},
		{
			name:           "sid with double quotes, value with no quotes",
			httpResponse:   `<html><input name="sid" value=foo></html>`,
			expectedOutput: ExpectedOutput{sid: "foo", err: nil}},
		{
			name:           "sid and value with single quotes",
			httpResponse:   `<html><input name='sid' value='foo'></html>`,
			expectedOutput: ExpectedOutput{sid: "foo", err: nil}},
		{
			name:           "sid with single quotes, value with double quotes",
			httpResponse:   `<html><input name='sid' value="foo"></html>`,
			expectedOutput: ExpectedOutput{sid: "foo", err: nil}},
		{
			name:           "sid with single quotes, value with no quotes",
			httpResponse:   `<html><input name='sid' value=foo></html>`,
			expectedOutput: ExpectedOutput{sid: "foo", err: nil}},
		{
			name:           "sid and value with no quotes",
			httpResponse:   `<html><input name=sid value=foo></html>`,
			expectedOutput: ExpectedOutput{sid: "foo", err: nil}},
		{
			name:           "sid with no quotes, value with double quotes",
			httpResponse:   `<html><input name=sid value=foo></html>`,
			expectedOutput: ExpectedOutput{sid: "foo", err: nil}},
		{
			name:           "sid with no quotes, value with single quotes",
			httpResponse:   `<html><input name=sid value='foo'></html>`,
			expectedOutput: ExpectedOutput{sid: "foo", err: nil}}}

	testDuo := Duo{httpClient: testClient}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpResponse = tt.httpResponse
			sid, err := testDuo.getSid("", "", "", testServerHostName)
			assert.Equal(tt.expectedOutput.sid, sid)
			assert.Equal(tt.expectedOutput.err, err)
		})
	}
}

func Test_Duo_checkMfaStatus(t *testing.T) {
	type ExpectedOutput struct {
		pushResponse *duoPushResponse
		err          error
	}
	assert := assert.New(t)

	tests := []struct {
		name           string
		httpResponse   string
		expectedOutput ExpectedOutput
	}{
		{
			name:           "non json response",
			httpResponse:   "this is not json",
			expectedOutput: ExpectedOutput{pushResponse: nil, err: ErrorJSONUnmarshalError}},
		{
			name:           "malformed json response",
			httpResponse:   `{"this": "jsonIs", "malformed"}`,
			expectedOutput: ExpectedOutput{pushResponse: nil, err: ErrorJSONUnmarshalError}},
		{
			name:           `happy path when checking user status after sending push`,
			httpResponse:   `{ "response": { "status_code": "pushed", "status": "Pushed a login request to your device..." }, "stat": "OK" }`,
			expectedOutput: ExpectedOutput{pushResponse: &duoPushResponse{Stat: "OK", Response: pushResponse{StatusCode: "pushed"}}, err: nil}},
		{
			name:           `happy path MFA accepted by user`,
			httpResponse:   `{ "response": { "reason": "User approved", "cookie": "AUTH|fakeAuthCookiePart1|fakeAuthCookiePart2", "result": "SUCCESS", "status": "Success. Logging you in...", "status_code": "allow", "parent": "https://parent.url.com/fake/path" }, "stat": "OK" }`,
			expectedOutput: ExpectedOutput{pushResponse: &duoPushResponse{Stat: "OK", Response: pushResponse{StatusCode: "allow", Parent: "https://parent.url.com/fake/path", Result: "SUCCESS", Cookie: "AUTH|fakeAuthCookiePart1|fakeAuthCookiePart2"}}, err: nil}}}

	testDuo := Duo{httpClient: testClient}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpResponse = tt.httpResponse
			pushResponse, err := testDuo.checkMfaStatus("", "", testServerHostName)
			assert.Equal(tt.expectedOutput.err, err)
			assert.EqualValues(tt.expectedOutput.pushResponse, pushResponse)
		})
	}
}
