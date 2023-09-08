package main

import (
	"log"
	"net/url"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/riotgames/key-conjurer/api/settings"
	"github.com/riotgames/key-conjurer/pkg/httputil"
)

func main() {
	settings, err := settings.NewSettings()
	if err != nil {
		log.Fatalf("Could not fetch configuration: %s", err)
	}

	oktaDomain := url.URL{
		Scheme: "https",
		Host:   settings.OktaHost,
	}

	service := keyconjurer.NewOktaService(&oktaDomain, settings.OktaToken)
	fn := httputil.Lambdaify(keyconjurer.ServeUserApplications(service))
	lambda.Start(fn)
}
