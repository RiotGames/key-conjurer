package httputil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

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

	req := http.Request{
		Method: inboundReq.HTTPMethod,
		URL: &url.URL{
			Path: inboundReq.Path,
		},
		Header: make(http.Header),
		Body:   io.NopCloser(strings.NewReader(inboundReq.Body)),
	}

	if len(inboundReq.MultiValueHeaders) > 0 {
		req.Header = http.Header(inboundReq.MultiValueHeaders)
	} else {
		for k, v := range inboundReq.Headers {
			req.Header[http.CanonicalHeaderKey(k)] = []string{v}
		}
	}

	var respWriter lambdaResponseWriter
	h.next.ServeHTTP(&respWriter, &req)
	if respWriter.TargetGroupResponse.StatusCode == 0 {
		respWriter.TargetGroupResponse.StatusCode = http.StatusOK
	}

	return json.Marshal(respWriter.TargetGroupResponse)
}

func Lambdaify(next http.Handler) lambda.Handler {
	return lambda2HttpHandler{next}
}
