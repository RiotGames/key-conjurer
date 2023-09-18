package main

import (
	"net/http"

	"golang.org/x/exp/slog"
)

type LogRoundTripper struct {
	RoundTripper http.RoundTripper
}

func FindOktaHeaders(r *http.Response) []slog.Attr {
	var attrs []slog.Attr
	if hdr := r.Header.Get("X-Okta-Request-Id"); hdr != "" {
		attrs = append(attrs, slog.String("okta_request_id", hdr))
	}
	return attrs
}

func (t LogRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	slog.Debug("HTTP Request", slog.String("url", r.URL.String()))
	resp, err := t.RoundTripper.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	// This array must be typed as any because slog.Debug() requires its arguments be any instead of slog.Attr.
	attrs := []any{
		slog.String("url", r.URL.String()),
		slog.Int("status_code", resp.StatusCode),
		slog.Bool("ok", resp.StatusCode == http.StatusOK),
	}

	for _, attr := range FindOktaHeaders(resp) {
		attrs = append(attrs, any(attr))
	}

	slog.Debug("HTTP Response", attrs...)
	return resp, nil
}
