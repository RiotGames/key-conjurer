package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/coreos/go-oidc"
	"github.com/riotgames/key-conjurer/internal/api"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := cli.Command{
		Action: runServer,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "okta-host",
				Usage:    "Okta host (e.g., 'example.okta.com')",
				Sources:  cli.EnvVars("KEYCONJURER_OKTA_HOST"),
				Required: true,
			},
			&cli.StringFlag{
				Name:    "okta-token",
				Sources: cli.EnvVars("KEYCONJURER_OKTA_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "okta-token-file",
				Sources: cli.EnvVars("KEYCONJURER_OKTA_TOKEN_FILE"),
			},
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func runServer(ctx context.Context, cmd *cli.Command) error {
	token := cmd.String("okta-token")
	if token == "" {
		tokenFile := cmd.String("okta-token-file")
		if tokenFile == "" {
			return cli.Exit("one of --okta-token or --okta-token-file must be specified", 1)
		}

		buf, err := os.ReadFile(tokenFile)
		if err != nil {
			return fmt.Errorf("could not read %s", tokenFile)
		}
		token = string(buf)
	}

	oktaDomain := url.URL{
		Scheme: "https",
		Host:   cmd.String("okta-host"),
	}

	service := api.NewOktaService(&oktaDomain, token)
	idp, err := oidc.NewProvider(ctx, oktaDomain.String())
	if err != nil {
		return fmt.Errorf("could not create OIDC provider: %w", err)
	}

	handler := api.ServeUserApplications(service, idp)
	lambda.StartWithOptions(handler, lambda.WithContext(ctx))
	return nil
}
