package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"golang.org/x/exp/slog"
)

type ApplicationType string

var (
	ApplicationTypeOIDC ApplicationType = "oidc"
	ApplicationTypeSAML ApplicationType = "saml"
)

type Application struct {
	ID   string          `json:"@id"`
	Name string          `json:"name"`
	Type ApplicationType `json:"type"`
	Href string          `json:"href"`
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
			ServeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		info, err := okta.GetUserInfo(ctx, idToken)
		if err != nil {
			if errors.Is(err, ErrBadRequest) {
				// Something went wrong within Okta, and the user can't do anything about it.
				slog.Error("okta indicated the request was poorly formed", requestAttrs...)
				ServeJSONError(w, http.StatusInternalServerError, "internal error when talking to the Okta API")
				return
			}

			requestAttrs = append(requestAttrs, slog.String("error", err.Error()))
			slog.Error("okta rejected id token", requestAttrs...)
			ServeJSONError(w, http.StatusForbidden, "unauthorized")
			return
		}

		requestAttrs = append(requestAttrs, slog.String("username", info.PreferredUsername))
		applications, err := okta.ListApplicationsForUser(ctx, info.PreferredUsername)
		if err != nil {
			requestAttrs = append(requestAttrs, slog.String("error", err.Error()))
			slog.Error("failed to fetch applications", requestAttrs...)
			ServeJSONError(w, http.StatusBadGateway, "upstream error")
			return
		}

		var accounts []Application
		for _, app := range applications {
			typ, ok := deriveApplicationType(app)
			if !ok {
				continue
			}

			app := Application{
				ID:   app.AppInstanceId,
				Name: app.Label,
				Type: typ,
				Href: app.LinkUrl,
			}

			accounts = append(accounts, app)
		}

		requestAttrs = append(requestAttrs, slog.Int("application_count", len(accounts)))
		slog.Info("served applications", requestAttrs...)
		ServeJSON(w, accounts)
	})
}

func deriveApplicationType(app *okta.AppLink) (ApplicationType, bool) {
	if app.AppName == "amazon_aws" {
		return ApplicationTypeOIDC, true
	}

	if strings.Contains(strings.ToLower(app.AppName), "tencent") {
		return ApplicationTypeSAML, true
	}

	return "", false
}
