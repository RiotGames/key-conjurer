package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/riotgames/key-conjurer/internal/api"
)

func main() {
	url, err := url.Parse(os.Getenv("OIDC_DOMAIN"))
	if err != nil {
		log.Fatalf("failed to parse OIDC_DOMAIN as URL: %s", url)
	}

	okta := api.NewOktaService(url, os.Getenv("OKTA_TOKEN"))
	http.Handle("/v2/applications", api.ServeUserApplications(okta))
	http.ListenAndServe(":8080", nil)
}
