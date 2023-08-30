package httputil

import "net/http"

type logRoundTripper struct {
	rt http.RoundTripper
}

func (t logRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return t.rt.RoundTrip(r)
}

func LogRoundTripper(rt http.RoundTripper) logRoundTripper {
	return logRoundTripper{rt}
}
