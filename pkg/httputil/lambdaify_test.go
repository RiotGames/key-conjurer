package httputil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetBearerToken(r *http.Request) (string, bool) {
	headerValue := r.Header.Get("authorization")
	if headerValue == "" {
		return "", false
	}

	parts := strings.Split(headerValue, " ")
	if len(parts) != 2 {
		return "", false
	}

	return parts[1], parts[0] == "Bearer"
}

func TestLambdaify_ALBTargetEvents(t *testing.T) {
	vals := url.Values{
		"id_token": {"id token goes here"},
	}

	inboundEvent := events.ALBTargetGroupRequest{
		HTTPMethod: "POST",
		Path:       "/hello-world",
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"authorization": "Bearer 1234",
		},
		Body: vals.Encode(),
	}

	inboundEventBytes, err := json.Marshal(inboundEvent)
	require.NoError(t, err, "Could not marshal inbound event to JSON")

	handler := Lambdaify(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST")
		assert.Equal(t, r.URL.Path, "/hello-world")
		assert.Equal(t, r.FormValue("id_token"), "id token goes here")
		token, ok := GetBearerToken(r)
		assert.Equal(t, token, "1234")
		assert.True(t, ok)

		w.Write([]byte("Hello, world!"))
	}))

	payload, err := handler.Invoke(context.Background(), []byte(inboundEventBytes))
	require.NoError(t, err, "Could not invoke Lambda handler")

	var resp events.ALBTargetGroupResponse
	require.NoError(t, json.Unmarshal(payload, &resp), "Could not unmarshal JSON")
	assert.Equal(t, resp.Body, "Hello, world!")
}
