package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

type OktaUserInfo struct {
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

// GetUserInfo returns user information about the given token
func (o Okta) GetUserInfo(ctx context.Context, token string) (info OktaUserInfo, err error) {
	if o.client == nil {
		o.client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, "GET", o.Domain.ResolveReference(&url.URL{Path: "/oauth2/v1/userinfo"}).String(), nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := o.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		buf, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(buf, &info)
		return
	case http.StatusUnauthorized:
		err = ErrUnauthorized
		return
	case http.StatusBadRequest:
		err = ErrBadRequest
		return
	}

	return
}
