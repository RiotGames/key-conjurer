package main

import (
	"encoding/json"
	"net/http"

	"github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/riotgames/key-conjurer/api/settings"
)

type server struct {
	h keyconjurer.Handler
}

func (s *server) getAWSCreds(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var ev keyconjurer.GetTemporaryCredentialEvent
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&ev); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := s.h.GetTemporaryCredentialEventHandler(r.Context(), ev)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	enc.Encode(resp)
}

func (s *server) getUserData(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var ev keyconjurer.GetUserDataEvent
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&ev); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := s.h.GetUserDataEventHandler(r.Context(), ev)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	enc.Encode(resp)
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

	enc := json.NewEncoder(w)
	enc.Encode(resp)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
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
