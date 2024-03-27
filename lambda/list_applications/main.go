package main

import (
	"context"
	"net/url"
	"os"

	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/riotgames/key-conjurer/internal"
	"github.com/riotgames/key-conjurer/internal/api"
)

func main() {
	settings, err := api.NewSettings(context.Background())
	if err != nil {
		slog.Error("could not fetch configuration: %s", err)
		os.Exit(1)
	}

	oktaDomain := url.URL{
		Scheme: "https",
		Host:   settings.OktaHost,
	}

	slog.Info("running list_applications_v2 Lambda")
	service := api.NewOktaService(&oktaDomain, settings.OktaToken)
	lambda.StartHandler(internal.Lambdaify(api.ServeUserApplications(service)))
}
