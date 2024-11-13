package main

import (
	"context"
	"net/url"
	"os"

	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/coreos/go-oidc"
	"github.com/riotgames/key-conjurer/internal/api"
)

func main() {
	settings, err := api.NewSettings(context.Background())
	if err != nil {
		slog.Error("could not fetch configuration: %s", "error", err)
		os.Exit(1)
	}

	oktaDomain := url.URL{
		Scheme: "https",
		Host:   settings.OktaHost,
	}

	service := api.NewOktaService(&oktaDomain, settings.OktaToken)
	idp, err := oidc.NewProvider(context.Background(), oktaDomain.String())
	if err != nil {
		slog.Error("could not create OIDC provider", "error", err)
		os.Exit(1)
	}

	slog.Info("running list_applications_v2 Lambda")

	lambda.Start(api.ServeUserApplications(service, idp))
}
