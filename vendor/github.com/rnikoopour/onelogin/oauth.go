package onelogin

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// OauthService handles communications with the authentication related methods on OneLogin.
type OauthService service

type authenticationParams struct {
	Username  string `json:"username_or_email"`
	Password  string `json:"password"`
	Subdomain string `json:"subdomain"`
}

type issueTokenParams struct {
	GrantType    string `json:"grant_type"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type getTokenResponse struct {
	AccessToken  string `json:"access_token"`
	AccountID    int    `json:"account_id"`
	CreatedAt    string `json:"created_at"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// An oauthToken authenticates request to OneLogin.
// It is valid for 3600 seconds, and can be renewed.
type oauthToken struct {
	AccessToken string
	AccountID   int
	CreatedAt   time.Time
	ExpiresIn   int64
	TokenType   string

	refreshToken string
	client       *Client
}

// isExpired check the OauthToken validity.
func (t *oauthToken) isExpired() bool {
	return time.Now().UTC().Add(-time.Second * time.Duration(t.ExpiresIn)).After(t.CreatedAt.UTC())
}

// refresh the token. The current token gets updates with new valid values.
func (t *oauthToken) refresh(ctx context.Context) error {
	u := "/auth/oauth2/token"
	b := issueTokenParams{
		GrantType:    "refresh_token",
		AccessToken:  t.AccessToken,
		RefreshToken: t.refreshToken,
	}
	req, err := t.client.NewRequest("POST", u, b)
	if err != nil {
		return err
	}

	var r []getTokenResponse
	_, err = t.client.Do(ctx, req, &r)
	if err != nil {
		return err
	}

	createdAt, _ := time.Parse(time.RFC3339Nano, r[0].CreatedAt)
	t.AccessToken = r[0].AccessToken
	t.AccountID = r[0].AccountID
	t.CreatedAt = createdAt
	t.ExpiresIn = r[0].ExpiresIn
	t.TokenType = r[0].TokenType
	t.refreshToken = r[0].RefreshToken

	return nil
}

// getToken issues a new token.
func (s *OauthService) getToken(ctx context.Context) (*oauthToken, error) {
	u := "/auth/oauth2/token"

	b := issueTokenParams{
		GrantType: "client_credentials",
	}
	req, err := s.client.NewRequest("POST", u, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("client_id: %s, client_secret: %s", s.client.clientID, s.client.clientSecret))

	var r []getTokenResponse
	_, err = s.client.Do(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	createdAt, _ := time.Parse(time.RFC3339Nano, r[0].CreatedAt)
	token := &oauthToken{
		AccessToken:  r[0].AccessToken,
		AccountID:    r[0].AccountID,
		CreatedAt:    createdAt,
		ExpiresIn:    r[0].ExpiresIn,
		TokenType:    r[0].TokenType,
		refreshToken: r[0].RefreshToken,
		client:       s.client,
	}

	return token, nil
}

type authenticateResponse struct {
	Status       string             `json:"status"`
	User         *AuthenticatedUser `json:"user"`
	ReturnToURL  string             `json:"return_to_url"`
	ExpiresAt    string             `json:"expires_at"`
	SessionToken string             `json:"session_token"`
	Devices      []Device           `json:"devices"`
}

// AuthenticatedUser contains user information for the Authentication.
type AuthenticatedUser struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	FirstName     string `json:"firstname"`
	LastName      string `json:"lastname"`
	Devices       []Device
	IsMfaRequired bool
}

func (u *AuthenticatedUser) SetMfaRequirement(required bool) {
	u.IsMfaRequired = required
}

func (u *AuthenticatedUser) SetDevices(devices []Device) {
	u.Devices = devices
}

// Authenticate a user from an email(or username) and a password.
func (s *OauthService) Authenticate(ctx context.Context, emailOrUsername string, password string) (*AuthenticatedUser, error) {
	u := "/api/1/login/auth"

	a := authenticationParams{
		Username:  emailOrUsername,
		Password:  password,
		Subdomain: s.client.subdomain,
	}

	req, err := s.client.NewRequest("POST", u, a)
	if err != nil {
		return nil, err
	}

	if err := s.client.AddAuthorization(ctx, req); err != nil {
		return nil, err
	}

	var d []authenticateResponse
	_, err = s.client.Do(ctx, req, &d)
	if err != nil {
		return nil, err
	}

	if len(d) != 1 || (d[0].Status != "Authenticated" && !d[0].User.IsMfaRequired) {
		return nil, errors.New("authentication failed")
	}

	return d[0].User, nil
}
