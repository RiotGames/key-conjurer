package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/okta/okta-sdk-golang/v2/okta"
)

func ServeUserApplications(client *okta.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

type lambdaResponseWriter struct {
	TargetGroupResponse events.ALBTargetGroupResponse
}

func (w *lambdaResponseWriter) Header() http.Header {
	if w.TargetGroupResponse.MultiValueHeaders == nil {
		w.TargetGroupResponse.MultiValueHeaders = make(map[string][]string)
	}

	return http.Header(w.TargetGroupResponse.MultiValueHeaders)
}

func (w *lambdaResponseWriter) WriteHeader(code int) {
	w.TargetGroupResponse.StatusCode = code
}

func (w *lambdaResponseWriter) Write(b []byte) (int, error) {
	w.TargetGroupResponse.Body = string(b)
	return len(b), nil
}

type lambda2HttpHandler struct {
	next http.Handler
}

func (h lambda2HttpHandler) Invoke(ctx context.Context, b []byte) ([]byte, error) {
	var inboundReq events.ALBTargetGroupRequest
	if err := json.Unmarshal(b, &inboundReq); err != nil {
		return nil, err
	}

	header := http.Header{}
	if len(inboundReq.MultiValueHeaders) > 0 {
		header = http.Header(inboundReq.MultiValueHeaders)
	} else {
		for k, v := range inboundReq.Headers {
			header[k] = []string{v}
		}
	}

	req := http.Request{
		Method: inboundReq.HTTPMethod,
		URL: &url.URL{
			Path: inboundReq.Path,
		},
		Header: header,
		Body:   io.NopCloser(strings.NewReader(inboundReq.Body)),
	}

	var respWriter lambdaResponseWriter
	h.next.ServeHTTP(&respWriter, &req)
	return json.Marshal(respWriter.TargetGroupResponse)
}

func lambdaify(next http.Handler) lambda.Handler {
	return lambda2HttpHandler{next}
}

func main() {
	var client okta.Client
	fn := lambdaify(ServeUserApplications(&client))
	lambda.Start(fn)
}
