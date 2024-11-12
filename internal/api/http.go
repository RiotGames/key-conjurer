package api

import (
	"strings"

	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"golang.org/x/oauth2"
)

// RequestAttrs returns attributes to be used with slog for the given request.
//
// []any is returned instead of []slog.Attr to make it easier to supply the attributes to slog functions using spread, for example:
//
//	slog.Error(msg, RequestAttrs(r)...)
func RequestAttrs(r events.ALBTargetGroupRequest) []any {
	var attrs []any

	if v, ok := r.Headers["x-amzn-trace-id"]; ok {
		attrs = append(attrs, slog.String("amz_request_id", v))
	}

	if v, ok := r.Headers["x-forwarded-for"]; ok {
		attrs = append(attrs, slog.String("x_forwarded_for", v))
	}

	return attrs
}

func requestTokenSource(r events.ALBTargetGroupRequest) (oauth2.TokenSource, bool) {
	headerValue, ok := r.Headers["authorization"]
	if !ok {
		return nil, false
	}

	parts := strings.Split(headerValue, " ")
	if len(parts) != 2 {
		return nil, false
	}
	if parts[0] != "Bearer" {
		return nil, false
	}

	token := oauth2.Token{AccessToken: parts[1]}
	return oauth2.StaticTokenSource(&token), true
}
