package api

import (
	"context"
	"errors"
	"net/http"

	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

type ServeUserApplicationsHandler struct {
	Okta OktaService
	Idp  *oidc.Provider
}

func (s ServeUserApplicationsHandler) Handle(ctx context.Context, r events.ALBTargetGroupRequest) (w events.ALBTargetGroupResponse, err error) {
	requestAttrs := RequestAttrs(r)
	ts, ok := requestTokenSource(r)
	if !ok {
		slog.Error("no bearer token present", requestAttrs...)
		ServeJSONError(&w, http.StatusUnauthorized, "unauthorized")
		return
	}

	info, err := s.Idp.UserInfo(ctx, ts)
	if err != nil {
		if errors.Is(err, ErrBadRequest) {
			slog.Error("okta indicated the request was poorly formed", requestAttrs...)
			ServeJSONError(&w, http.StatusInternalServerError, "internal error when talking to the Okta API")
			return
		}

		requestAttrs = append(requestAttrs, slog.String("error", err.Error()))
		slog.Error("okta rejected id token", requestAttrs...)
		ServeJSONError(&w, http.StatusForbidden, "unauthorized")
		return
	}

	var claims Claims
	if err := info.Claims(&claims); err != nil {
		slog.Error("failed to parse claims from Okta userinfo endpoint", requestAttrs...)
		ServeJSONError(&w, http.StatusInternalServerError, "internal error when talking to the Okta API")
	}

	requestAttrs = append(requestAttrs, slog.String("username", claims.PreferredUsername))
	applications, err := s.Okta.ListApplicationsForUser(ctx, claims.PreferredUsername)
	if err != nil {
		requestAttrs = append(requestAttrs, slog.String("error", err.Error()))
		slog.Error("failed to fetch applications", requestAttrs...)
		ServeJSONError(&w, http.StatusBadGateway, "upstream error")
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
	ServeJSON(&w, accounts)
	return
}

func (s ServeUserApplicationsHandler) Handler() lambda.Handler {
	return lambda.NewHandler(s.Handle)
}

func ServeUserApplications(okta OktaService, idp *oidc.Provider) lambda.Handler {
	h := ServeUserApplicationsHandler{
		Okta: okta,
		Idp:  idp,
	}

	return h.Handler()
}
