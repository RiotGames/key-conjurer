package main

import (
	"context"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/riotgames/key-conjurer/api/keyconjurer"
	"github.com/riotgames/key-conjurer/pkg/httputil"
	"golang.org/x/exp/slog"
)

func main() {
	settings, err := keyconjurer.NewSettings(context.Background())
	if err != nil {
		slog.Error("could not fetch configuration: %s", err)
		os.Exit(1)
	}

	oktaDomain := url.URL{
		Scheme: "https",
		Host:   settings.OktaHost,
	}

	slog.Info("running list_applications_v2 Lambda")
	service := keyconjurer.NewOktaService(&oktaDomain, settings.OktaToken)
	lambda.StartHandler(httputil.Lambdaify(keyconjurer.ServeUserApplications(service)))
}
