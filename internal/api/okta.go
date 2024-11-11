package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/okta/okta-sdk-golang/v2/okta"
)

type Okta struct {
	Domain     *url.URL
	Token      string
	client     *http.Client
	oktaClient *okta.Client
}

func NewOktaService(domain *url.URL, token string) Okta {
	_, oktaClient, _ := okta.NewClient(
		context.Background(),
		okta.WithToken(token),
		okta.WithOrgUrl(domain.String()),
	)

	return Okta{domain, token, http.DefaultClient, oktaClient}
}

func (o Okta) ListApplicationsForUser(ctx context.Context, user string) ([]*okta.AppLink, error) {
	links, resp, err := o.oktaClient.User.ListAppLinks(ctx, user)
	if err != nil {
		return nil, err
	}

	for resp.HasNextPage() {
		var next []*okta.AppLink
		if resp, err = resp.Next(ctx, &next); err != nil {
			return nil, err
		}

		links = append(links, next...)
	}

	return links, nil
}

type Claims struct {
	Sub               string `json:"sub"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	ZoneInfo          string `json:"zoneinfo"`
	Locale            string `json:"locale"`
}

var (
	ErrBadRequest   = errors.New("bad request")
	ErrUnauthorized = errors.New("unauthorized")
)
