package onelogin

import (
	"context"
	"strings"
)

type SAMLService service

type samlRequestParams struct {
	OtpToken   string `json:"otp_token"`
	DeviceID   string `json:"device_id"`
	AppID      string `json:"app_id"`
	StateToken string `json:"state_token"`
}

type stateTokenParams struct {
	Username  string `json:"username_or_email"`
	Password  string `json:"password"`
	AppID     string `json:"app_id"`
	Subdomain string `json:"subdomain"`
}

type Device struct {
	Id               int64  `json:"device_id"`
	DeviceType       string `json:"device_type"`
	ApiHostName      string `json:"duo_api_hostname"`
	SignatureRequest string `json:"duo_sig_request"`
}

type SamlUser struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	FirstName     string `json:"firstname"`
	LastName      string `json:"lastname"`
	IsMfaRequired bool
}

func (u *SamlUser) SetMfaRequirement(required bool) {
	u.IsMfaRequired = required
}

type MFAResponse struct {
	StateToken  string   `json:"state_token"`
	User        SamlUser `json:"user"`
	Devices     []Device `json:"devices"`
	CallbackUrl string   `json:"callback_url"`
}

type SamlResponse struct {
	SamlString string
}

func (s *SamlResponse) SetSamlString(saml string) {
	s.SamlString = saml
}

func (s *SAMLService) SamlAssertion(ctx context.Context, username, password, appID string) (*MFAResponse, error) {
	u := "/api/1/saml_assertion"
	a := stateTokenParams{
		Username:  username,
		Password:  password,
		AppID:     appID,
		Subdomain: s.client.subdomain}

	req, err := s.client.NewRequest("POST", u, a)
	mfaResponse := []MFAResponse{}
	if err != nil {
		return nil, err
	}

	if err := s.client.AddAuthorization(ctx, req); err != nil {
		return nil, err
	}
	if _, err := s.client.Do(ctx, req, &mfaResponse); err != nil {
		return nil, err
	}
	return &mfaResponse[0], nil
}

func (s *SAMLService) VerifyFactor(ctx context.Context, otp, stateToken, appId, deviceId string) (string, error) {
	u := "/api/1/saml_assertion/verify_factor"
	a := samlRequestParams{
		OtpToken:   otp,
		DeviceID:   deviceId,
		AppID:      appId,
		StateToken: stateToken}
	req, err := s.client.NewRequest("POST", u, a)
	if err != nil {
		return "", err
	}

	if err := s.client.AddAuthorization(ctx, req); err != nil {
		return "", err
	}
	samlResponse := &SamlResponse{}
	if _, err := s.client.Do(ctx, req, samlResponse); err != nil {
		return "", err
	}
	// Need to remove the double quote artifact from converting a json.RawMessage
	//  into a Go string
	return strings.Trim(samlResponse.SamlString, "\""), nil
}
