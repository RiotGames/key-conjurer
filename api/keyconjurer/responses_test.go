package keyconjurer

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type T struct {
	Foo, Bar string
}

func TestDataResponse(t *testing.T) {
	data := T{Foo: "Foo", Bar: "Qux"}
	proxyResponse, err := DataResponse(data)
	require.NoError(t, err, "no error should be returned")
	require.NotNil(t, proxyResponse, "data response should not be nil")
	require.Equal(t, 200, proxyResponse.StatusCode, "unexpected status code")
	require.Equal(t, "application/json", proxyResponse.Headers["Content-Type"], "unexpeted content type")
	require.Equal(t, "*", proxyResponse.Headers["Access-Control-Allow-Origin"], "unexpected CORS header")

	var response Response
	response.UnmarshalJSON([]byte(proxyResponse.Body))
	require.True(t, response.Success)
	require.Equal(t, "success", response.Message)
}

func TestErrorResponse(t *testing.T) {
	message := "wrong credentials"
	proxyResponse, err := ErrorResponse(ErrCodeInvalidCredentials, message)
	require.NoError(t, err)
	require.NotNil(t, proxyResponse)
	require.Equal(t, 403, proxyResponse.StatusCode, "unexpected status code")
	require.Equal(t, "application/json", proxyResponse.Headers["Content-Type"], "unexpeted content type")
	require.Equal(t, "*", proxyResponse.Headers["Access-Control-Allow-Origin"], "unexpected CORS header")

	var response Response
	response.UnmarshalJSON([]byte(proxyResponse.Body))
	require.False(t, response.Success)
	require.Equal(t, message, response.Message)
}

func TestErrorResponseStatusCodes(t *testing.T) {
	proxyResponse, err := ErrorResponse(ErrCodeBadRequest, "bad request")
	require.NoError(t, err)
	require.NotNil(t, proxyResponse)
	require.Equal(t, 400, proxyResponse.StatusCode, "unexpected status code")

	proxyResponse, err = ErrorResponse(ErrCodeInternalServerError, "bad request")
	require.NoError(t, err)
	require.NotNil(t, proxyResponse)
	require.Equal(t, 500, proxyResponse.StatusCode, "unexpected status code")
}

func TestResponseMarshalJSON(t *testing.T) {
	response, err := DataResponse(T{Foo: "Foo", Bar: "Qux"})
	require.NoError(t, err)

	b, err := json.Marshal(response)
	require.NoError(t, err)

	expectedBody := `{\"Success\":true,\"Message\":\"success\",\"Data\":{\"Foo\":\"Foo\",\"Bar\":\"Qux\"}}`
	expectedHeaders := `{"Access-Control-Allow-Origin":"*","Content-Type":"application/json"}`
	expectedData := fmt.Sprintf(`{"statusCode":200,"headers":%s,"multiValueHeaders":{},"body":"%s"}`, expectedHeaders, expectedBody)
	require.Equal(t, expectedData, string(b))
}

func TestErrorResponseMarshalJSON(t *testing.T) {
	message := "this is a error message"

	response, err := ErrorResponse(ErrCodeBadRequest, message)
	require.NoError(t, err)
	require.NotNil(t, response)

	b, err := json.Marshal(response)
	require.NoError(t, err)
	require.NotNil(t, b)

	expectedBody := fmt.Sprintf(`{\"Success\":false,\"Message\":\"%s\",\"Data\":{\"Code\":\"bad_request\",\"Message\":\"%s\"}}`, message, message)
	expectedHeaders := `{"Access-Control-Allow-Origin":"*","Content-Type":"application/json"}`
	expectedData := fmt.Sprintf(`{"statusCode":400,"headers":%s,"multiValueHeaders":{},"body":"%s"}`, expectedHeaders, expectedBody)
	require.Equal(t, expectedData, string(b))
}

func TestResponseGetPayload(t *testing.T) {
	payload := `{"Success":true,"Message":"","Data":{"foo": "bar", "qux": "baz"}}`
	var response Response
	var data map[string]string
	var err ErrorData
	require.Error(t, response.GetPayload(&data))
	require.Error(t, response.GetError(&err))
	require.NoError(t, json.Unmarshal([]byte(payload), &response))
	require.NoError(t, response.GetPayload(&data))
	require.Error(t, response.GetError(&err))
	require.Equal(t, "bar", data["foo"])
	require.Equal(t, "baz", data["qux"])
}

func TestResponseGetError(t *testing.T) {
	payload := `{"Success":false,"Data":{"Code": "unspecified", "Message": "Something broke"}}`
	var response Response
	var data map[string]string
	var err ErrorData
	require.Error(t, response.GetPayload(&data))
	require.NoError(t, json.Unmarshal([]byte(payload), &response))
	require.Error(t, response.GetPayload(&data))
	require.NoError(t, response.GetError(&err))
	require.Equal(t, "Something broke (code: unspecified)", err.Error())
}
