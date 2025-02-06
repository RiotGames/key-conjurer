package oauth2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

// ErrBadRequest indicates that the request to the oauth2 callback endpoint contained malformed data.
var ErrBadRequest = errors.New("bad request")

type job struct {
	State, Verifier string
	C               chan *oauth2.Token
}

type codeExchanger interface {
	Exchange(ctx context.Context, code string, params ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type handler struct {
	Exchanger     codeExchanger
	jobs          chan job
	serveResponse func(err error, w http.ResponseWriter, r *http.Request) error
}

func defaultResponseHandler(err error, w http.ResponseWriter, _ *http.Request) error {
	if err == nil {
		fmt.Fprintln(w, "You may close this window now.")
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		w.WriteHeader(http.StatusGatewayTimeout)
		fmt.Fprintln(w, "Took too long to get credentials.")
		return nil
	}

	if errors.Is(err, ErrBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Bad request.")
		return nil
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "Internal server error.")
	return nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var job job
	srvResponse := h.serveResponse
	if srvResponse == nil {
		srvResponse = defaultResponseHandler
	}

	ctx := r.Context()
	select {
	case j := <-h.jobs:
		job = j
		defer close(job.C)
	case <-ctx.Done():
		srvResponse(ctx.Err(), w, r)
		return
	}

	req := authorizationCodeReq{
		errorMessage:     r.FormValue("error"),
		errorDescription: r.FormValue("error_description"),
		state:            r.FormValue("state"),
		code:             r.FormValue("code"),
	}

	var code string
	if err := req.Verify(req.state, &code); err != nil {
		srvResponse(err, w, r)
		return
	}
	token, err := h.Exchanger.Exchange(ctx, code, oauth2.VerifierOption(job.Verifier))
	if err != nil {
		srvResponse(err, w, r)
		return
	}

	select {
	case job.C <- token:
		srvResponse(nil, w, r)
	case <-ctx.Done():
		srvResponse(ctx.Err(), w, r)
	}
}

func (h *handler) Wait(ctx context.Context, state, verifier string) (*oauth2.Token, error) {
	j := job{State: state, Verifier: verifier, C: make(chan *oauth2.Token)}
	select {
	case h.jobs <- j:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case tok := <-j.C:
		return tok, nil
	}
}

func (h *handler) Close() error {
	close(h.jobs)
	return nil
}

type authorizationCodeReq struct {
	code             string
	state            string
	errorMessage     string
	errorDescription string
}

func (o authorizationCodeReq) Verify(state string, code *string) error {
	if o.errorMessage != "" || strings.Compare(o.state, state) != 0 {
		return ErrBadRequest
	}
	*code = o.code
	return nil
}
