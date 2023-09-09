package keyconjurer

import (
	"net/http"
	"strings"

	"golang.org/x/exp/slog"
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

	if v, ok := r.Header["X-Amzn-Trace-Id"]; ok {
		attrs = append(attrs, slog.String("amz_request_id", v[0]))
	}

	if v, ok := r.Header["X-Forwarded-For"]; ok {
		attrs = append(attrs, slog.String("x_forwarded_for", v[0]))
	}

	return attrs
}

func GetBearerToken(r *http.Request) (string, bool) {
	headerValue, ok := r.Header["Authorization"]
	if !ok {
		return "", false
	}

	if len(headerValue) != 1 {
		return "", false
	}

	parts := strings.Split(headerValue[0], " ")
	if len(parts) != 2 {
		return "", false
	}

	return parts[1], parts[0] == "Bearer"
}
