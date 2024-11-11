package api

import (
	"context"
	"errors"
	"net/http"

	"log/slog"

	"github.com/coreos/go-oidc"
	"github.com/okta/okta-sdk-golang/v2/okta"
)

type Application struct {
	ID   string `json:"@id"`
	Name string `json:"name"`
}

type OktaService interface {
	ListApplicationsForUser(ctx context.Context, user string) ([]*okta.AppLink, error)
}

func ServeUserApplications(okta OktaService, idp *oidc.Provider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestAttrs := RequestAttrs(r)
		ts, ok := requestTokenSource(r)
		if !ok {
			slog.Error("no bearer token present", requestAttrs...)
			ServeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		info, err := idp.UserInfo(ctx, ts)
		if err != nil {
			if errors.Is(err, ErrBadRequest) {
				slog.Error("okta indicated the request was poorly formed", requestAttrs...)
				ServeJSONError(w, http.StatusInternalServerError, "internal error when talking to the Okta API")
				return
			}

			requestAttrs = append(requestAttrs, slog.String("error", err.Error()))
			slog.Error("okta rejected id token", requestAttrs...)
			ServeJSONError(w, http.StatusForbidden, "unauthorized")
			return
		}

		var claims Claims
		if err := info.Claims(&claims); err != nil {
			slog.Error("failed to parse claims from Okta userinfo endpoint", requestAttrs...)
			ServeJSONError(w, http.StatusInternalServerError, "internal error when talking to the Okta API")
		}

		requestAttrs = append(requestAttrs, slog.String("username", claims.PreferredUsername))
		applications, err := okta.ListApplicationsForUser(ctx, claims.PreferredUsername)
		if err != nil {
			requestAttrs = append(requestAttrs, slog.String("error", err.Error()))
			slog.Error("failed to fetch applications", requestAttrs...)
			ServeJSONError(w, http.StatusBadGateway, "upstream error")
			return
		}

		var accounts []Application
		for _, app := range applications {
			if app.AppName == "amazon_aws" {
				accounts = append(accounts, Application{
					ID:   app.AppInstanceId,
					Name: app.Label,
				})
			}
		}

		requestAttrs = append(requestAttrs, slog.Int("application_count", len(accounts)))
		slog.Info("served applications", requestAttrs...)
		ServeJSON(w, accounts)
	})
}
