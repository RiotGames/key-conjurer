package keyconjurer

import (
	"context"
	"net/http"
	"strings"

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

	return parts[1], parts[0] == "Bearer"
}

type OktaService interface {
	GetUserInfo(ctx context.Context, token string) (OktaUserInfo, error)
	ListApplicationsForUser(ctx context.Context, user string) ([]*okta.AppLink, error)
}

func ServeUserApplications(okta OktaService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attrs := []any{
			slog.String("origin_ip_address", r.RemoteAddr),
		}

		if v, ok := r.Header["X-Amzn-Trace-Id"]; ok {
			attrs = append(attrs, slog.String("amz_request_id", v[0]))
		}

		if v, ok := r.Header["X-Forwarded-For"]; ok {
			attrs = append(attrs, slog.String("x_forwarded_for", v[0]))
		}

		idToken, ok := GetBearerToken(r)
		if !ok {
			slog.Error("no bearer token present", attrs...)
			httputil.ServeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		info, err := okta.GetUserInfo(r.Context(), idToken)
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			slog.Error("okta rejected id token", attrs...)
			httputil.ServeJSONError(w, http.StatusForbidden, "unauthorized")
			return
		}

		attrs = append(attrs, slog.String("username", info.PreferredUsername))
		applications, err := okta.ListApplicationsForUser(r.Context(), info.PreferredUsername)
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			slog.Error("failed to fetch applications", attrs...)
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

		attrs = append(attrs, slog.Int("application_count", len(accounts)))
		slog.Info("served applications", attrs...)
		httputil.ServeJSON(w, accounts)
	})
}
