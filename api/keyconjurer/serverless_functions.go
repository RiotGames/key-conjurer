package keyconjurer

import (
	"context"
	"net/http"
	"strings"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/riotgames/key-conjurer/internal/api"
	"golang.org/x/exp/slog"
)

type Application struct {
	ID   string `json:"@id"`
	Name string `json:"name"`
}

type OktaService interface {
	GetUserInfo(ctx context.Context, token string) (OktaUserInfo, error)
	ListApplicationsForUser(ctx context.Context, user string) ([]*okta.AppLink, error)
}

func ServeUserApplications(okta OktaService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestAttrs := RequestAttrs(r)
		idToken, ok := GetBearerToken(r)
		if !ok {
			slog.Error("no bearer token present", requestAttrs...)
			api.ServeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		info, err := okta.GetUserInfo(ctx, idToken)
		if err != nil {
			requestAttrs = append(requestAttrs, slog.String("error", err.Error()))
			slog.Error("okta rejected id token", requestAttrs...)
			api.ServeJSONError(w, http.StatusForbidden, "unauthorized")
			return
		}

		requestAttrs = append(requestAttrs, slog.String("username", info.PreferredUsername))
		applications, err := okta.ListApplicationsForUser(ctx, info.PreferredUsername)
		if err != nil {
			requestAttrs = append(requestAttrs, slog.String("error", err.Error()))
			slog.Error("failed to fetch applications", requestAttrs...)
			api.ServeJSONError(w, http.StatusBadGateway, "upstream error")
			return
		}

		var accounts []Application
		for _, app := range applications {
			if app.AppName == "amazon_aws" || strings.Contains(app.AppName, "tencent") {
				accounts = append(accounts, Application{
					ID:   app.AppInstanceId,
					Name: app.Label,
				})
			}
		}

		requestAttrs = append(requestAttrs, slog.Int("application_count", len(accounts)))
		slog.Info("served applications", requestAttrs...)
		api.ServeJSON(w, accounts)
	})
}
