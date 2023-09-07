package keyconjurer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/okta/okta-sdk-golang/v2/okta"
)

type oktaService struct {
	Domain *url.URL
	Token  string
	client *http.Client
}

func NewOktaService(domain *url.URL, token string) oktaService {
	return oktaService{domain, token, http.DefaultClient}
}

func (oktaService) ListApplicationsForUser(ctx context.Context, user string) ([]okta.Application, error) {
	return nil, nil
}

type OktaUserInfo struct {
	Sub               string    `json:"sub"`
	GivenName         string    `json:"given_name"`
	FamilyName        string    `json:"family_name"`
	PreferredUsername string    `json:"preferred_username"`
	Email             string    `json:"email"`
	ZoneInfo          string    `json:"zoneinfo"`
	Locale            string    `json:"locale"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// GetUserInfo returns user information about the given token
func (o oktaService) GetUserInfo(ctx context.Context, token string) (info OktaUserInfo, err error) {
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
		err = fmt.Errorf("invalid token: %w", err)
		return
	}
	defer resp.Body.Close()

	buf, _ := io.ReadAll(resp.Body)
	json.Unmarshal(buf, &info)
	return
}
