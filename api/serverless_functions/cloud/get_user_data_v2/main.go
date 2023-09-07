package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/pkg/httputil"
	"golang.org/x/exp/slog"
)

func GetBearerToken(r *http.Request) (string, bool) {
	headerValue, ok := r.Header["Authorization"]
	if !ok {
		return "", false
	}

	if len(headerValue) != 1 {
		return "", false
	}

	parts := strings.Split(headerValue[0], " ")
	if len(parts) != 2 {
		return "", false
	}

	return parts[2], parts[1] == "Bearer"
}

type OktaService interface {
	// GetTokenUsername returns the username associated with the given id token, or an error if the token is not accepted by Okta for any reason.
	GetTokenUsername(ctx context.Context, token string) (string, error)
	ListApplicationsForUser(ctx context.Context, user string) ([]okta.Application, error)
}

func ServeUserApplications(okta OktaService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originIpAddr := r.Header.Get("X-Forwarded-For")
		idToken, ok := GetBearerToken(r)
		if !ok {
			slog.Error("no bearer token present",
				slog.String("origin_ip_address", originIpAddr),
			)

			httputil.ServeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		username, err := okta.GetTokenUsername(r.Context(), idToken)
		if err != nil {
			slog.Error("okta rejected id token",
				slog.String("error", err.Error()),
				slog.String("origin_ip_address", originIpAddr),
			)

			httputil.ServeJSONError(w, http.StatusForbidden, "unauthorized")
			return
		}

		applications, err := okta.ListApplicationsForUser(r.Context(), username)
		if err != nil {
			slog.Error("failed to fetch applications",
				slog.String("error", err.Error()),
				slog.String("origin_ip_address", originIpAddr),
			)

			httputil.ServeJSONError(w, http.StatusBadGateway, "upstream error")
			return
		}

		accounts := make([]core.Application, len(applications))
		for i, app := range applications {
			accounts[i] = core.Application{
				ID:   app.Id,
				Name: app.Label,
			}
		}

		slog.Info("served applications",
			slog.Int("application_count", len(accounts)),
			slog.String("origin_ip_address", originIpAddr),
		)

		httputil.ServeJSON(w, accounts)
	})
}

func main() {
	var service OktaService
	fn := httputil.Lambdaify(ServeUserApplications(service))
	lambda.Start(fn)
}
