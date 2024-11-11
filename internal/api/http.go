package api

import (
	"net/http"
	"strings"

	"log/slog"

	"golang.org/x/oauth2"
)

// RequestAttrs returns attributes to be used with slog for the given request.
//
// []any is returned instead of []slog.Attr to make it easier to supply the attributes to slog functions using spread, for example:
//
//	slog.Error(msg, RequestAttrs(r)...)
func RequestAttrs(r *http.Request) []any {
	attrs := []any{
		slog.String("origin_ip_address", r.RemoteAddr),
	}

	if v := r.Header.Get("x-amzn-trace-id"); v != "" {
		attrs = append(attrs, slog.String("amz_request_id", v))
	}

	if v := r.Header.Get("x-forwarded-for"); v != "" {
		attrs = append(attrs, slog.String("x_forwarded_for", v))
	}

	return attrs
}

func requestTokenSource(r *http.Request) (oauth2.TokenSource, bool) {
	headerValue := r.Header.Get("authorization")
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
