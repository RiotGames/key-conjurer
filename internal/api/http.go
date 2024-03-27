package api

import (
	"net/http"
	"strings"

	"log/slog"
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
