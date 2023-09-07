package main

import (
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/riotgames/key-conjurer/api/settings"
)

var (
	OktaToken  = os.Getenv("OKTA_TOKEN")
	OktaDomain *url.URL
)

func init() {
	uri, err := url.Parse(os.Getenv("OKTA_DOMAIN"))
	if err != nil {
		log.Fatalf("OKTA_DOMAIN must be a valid URL: %s", err)
	}
	OktaDomain = uri
}

type server struct {
	h keyconjurer.Handler
}

// encodeTargetGroupResponse encodes the given ALBTargetGroupResponse to JSON
//
// In normal operation, AWS will extract our payload from this response.
// We must manually do this in the test server because the client will not understand it.
func encodeTargetGroupResponse(w http.ResponseWriter, response *events.ALBTargetGroupResponse) {
	w.Header().Set("Content-Type", mime.FormatMediaType("application/json", map[string]string{"encoding": "utf8"}))

	// Body is already JSON
	w.Write([]byte(response.Body))
}

func (s *server) getAWSCreds(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request := &events.ALBTargetGroupRequest{Body: string(body)}
	resp, err := s.h.GetTemporaryCredentialEventHandler(r.Context(), request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encodeTargetGroupResponse(w, resp)
}

func (s *server) getUserData(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	request := &events.ALBTargetGroupRequest{Body: string(body)}
	resp, err := s.h.GetUserDataEventHandler(r.Context(), request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encodeTargetGroupResponse(w, resp)
}

func (s *server) listAuthenticationProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	resp, err := s.h.ListProvidersHandler(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encodeTargetGroupResponse(w, resp)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
	}
	switch r.URL.Path {
	case "/v2/applications":
		okta := keyconjurer.NewOktaService(
			OktaDomain,
			OktaToken,
		)
		handler := keyconjurer.ServeUserApplications(&okta)
		handler.ServeHTTP(w, r)
	case "/get_aws_creds":
		s.getAWSCreds(w, r)
	case "/get_user_data":
		s.getUserData(w, r)
	case "/list_providers":
		s.listAuthenticationProviders(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	cfg, err := settings.NewSettings()
	if err != nil {
		panic(err)
	}

	s := server{h: keyconjurer.NewHandler(cfg)}
	http.ListenAndServe("127.0.0.1:4000", &s)
}
