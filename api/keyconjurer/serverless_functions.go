package keyconjurer

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/riotgames/key-conjurer/api/authenticators/duo"
	"github.com/riotgames/key-conjurer/api/authenticators/okta"
	onelogin "github.com/riotgames/key-conjurer/api/authenticators/onelogin_duo"
	"github.com/riotgames/key-conjurer/api/aws"
	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/api/settings"
)

type Handler struct {
	crypt                   core.Crypto
	cfg                     *settings.Settings
	aws                     *aws.Provider
	authenticationProviders providerMap
}

func NewHandler(cfg *settings.Settings) Handler {
	client, err := aws.NewProvider(cfg.AwsRegion)
	if err != nil {
		// TODO Probably shouldn't be a panic
		panic(err)
	}

	mfa := duo.New()
	var prov core.CryptoProvider = &core.PassThroughProvider{}
	if cfg.AwsKMSKeyID != "" {
		prov = core.NewKMSProvider(&core.KMSProviderConfig{
			KMSKeyID: cfg.AwsKMSKeyID,
			// TODO: I think this might cause issues if the KMS secret is not in the same region as the Lambda
			Session: session.New(),
		})
	}

	return Handler{
		crypt: core.NewCrypto(prov),
		cfg:   cfg,
		aws:   client,
		authenticationProviders: providerMap{
			AuthenticationProviderOkta:     okta.Must(cfg.OktaHost, cfg.OktaToken, mfa),
			AuthenticationProviderOneLogin: onelogin.New(cfg, mfa),
		},
	}
}

type GetUserDataEvent struct {
	core.Credentials
	// AuthenticationProvider is the authentication provider that should be used when logging in.
	AuthenticationProvider AuthenticationProviderName `json:"authentication_provider"`
}

type GetUserDataPayload struct {
	Apps                 []core.Application `json:"apps"`
	EncryptedCredentials string             `json:"creds"`
}

// GetUserDataEventHandler authenticates the user against OneLogin and retrieves a list of AWS application the user has available.
//
// This MUST be backwards compatible with the old version of KeyConjurer for a time.
func (h *Handler) GetUserDataEventHandler(ctx context.Context, event GetUserDataEvent) (Response, error) {
	creds := event.Credentials
	provider, ok := h.authenticationProviders.Get(event.AuthenticationProvider)
	if !ok {
		return ErrorResponse(ErrCodeInvalidProvider, "the provider you supplied is unsupported by this version of KeyConjurer")
	}

	if err := h.crypt.Decrypt(ctx, &creds); err != nil {
		return ErrorResponse(ErrCodeUnableToDecrypt, "unable to decrypt credentials")
	}

	user, err := provider.Authenticate(ctx, creds)
	if err != nil {
		// TODO: provide more detailed errors - this could fail because of an upstream error (provider being down) or because of an error with the users credentials
		return ErrorResponse(ErrCodeInvalidCredentials, "credentials are incorrect")
	}

	applications, err := provider.ListApplications(ctx, user)
	if err != nil {
		return ErrorResponse(ErrCodeInternalServerError, fmt.Sprintf("failed to retrieve applications: %s", err))
	}

	ciphertext, err := h.crypt.Encrypt(ctx, creds)
	if err != nil {
		return ErrorResponse(ErrCodeUnableToEncrypt, "unable to encrypt credentials")
	}

	return DataResponse(GetUserDataPayload{
		Apps:                 applications,
		EncryptedCredentials: ciphertext,
	})
}

type GetTemporaryCredentialEvent struct {
	core.Credentials
	AppID          string `json:"appId"`
	TimeoutInHours uint8  `json:"timeoutInHours"`
	RoleName       string `json:"roleName"`

	// AuthenticationProvider is the authentication provider that should be used when logging in.
	// This will be blank for old versions of KeyConjurer; if it is blank, you must default to OneLogin
	AuthenticationProvider AuthenticationProviderName `json:"authentication_provider"`
}

var (
	errTimeoutBadSize = errors.New("ttl must be at least 1 hour and less than 8 hours")
	errNoRoleProvided = errors.New("a role must be specified when using this identity provider")
)

// Validate validates that the event has appropriate parameters
func (e *GetTemporaryCredentialEvent) Validate() error {
	if e.TimeoutInHours < 1 || e.TimeoutInHours > 8 {
		return errTimeoutBadSize
	}

	if e.AuthenticationProvider == AuthenticationProviderOneLogin {
		// We don't use role names in OneLogin
		e.RoleName = ""
	}

	if e.AuthenticationProvider == AuthenticationProviderOkta && e.RoleName == "" {
		return errNoRoleProvided
	}

	return nil
}

type GetTemporaryCredentialsPayload struct {
	// TODO: add CloudProvider property so the client can discriminate between different cloud providers
	AccountID       string `json:"AccountId"` // Intentionally lower-cased to maintain backwards compatibilty
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

// GetTemporaryCredentialEventHandler issues temporary credentials for the current user.
//
// This MUST be backwards compatible with the old version of KeyConjurer for a time.
func (h *Handler) GetTemporaryCredentialEventHandler(ctx context.Context, event GetTemporaryCredentialEvent) (Response, error) {
	if err := event.Validate(); err != nil {
		return ErrorResponse(ErrBadRequest, err.Error())
	}

	creds := event.Credentials
	provider, ok := h.authenticationProviders.Get(event.AuthenticationProvider)
	if !ok {
		return ErrorResponse(ErrCodeInvalidProvider, "invalid provider")
	}

	if err := h.crypt.Decrypt(ctx, &creds); err != nil {
		return ErrorResponse(ErrCodeUnableToDecrypt, "unable to decrypt credentials")
	}

	response, err := provider.GenerateSAMLAssertion(ctx, creds, event.AppID)
	if err != nil {
		msg := fmt.Sprintf("unable to generate SAML assertion: %s", err)
		return ErrorResponse(ErrCodeInternalServerError, msg)
	}

	sts, err := h.aws.GetTemporaryCredentialsForUser(ctx, event.RoleName, response, int(event.TimeoutInHours))
	if err != nil {
		var errRoleNotFound aws.ErrRoleNotFound
		if errors.As(err, &errRoleNotFound) {
			return ErrorResponse(ErrBadRequest, errRoleNotFound.Error())
		}

		return ErrorResponse(ErrCodeInternalServerError, err.Error())
	}

	return DataResponse(GetTemporaryCredentialsPayload{
		AccountID:       event.AppID,
		AccessKeyID:     *sts.AccessKeyID,
		SecretAccessKey: *sts.SecretAccessKey,
		SessionToken:    *sts.SessionToken,
		Expiration:      sts.Expiration,
	})
}

// ListProvidersEvent is the set of parameters available for listing providers.
type ListProvidersEvent struct {
	// This is intentionally an empty struct; there are no parameters for listing authentication providers
	// It is used to keep the Go API consistent with other endpoints.
}

type Provider struct {
	ID string
}

type ListProvidersPayload struct {
	Providers []Provider
}

// ListProvidersHandler allows a user to list the providers they may authenticate with.
//
// This does NOT need to be backwards compatible with old KeyConjurer clients.
func (h *Handler) ListProvidersHandler(ctx context.Context) (Response, error) {
	var p []Provider
	for key := range h.authenticationProviders {
		p = append(p, Provider{ID: key})
	}

	return DataResponse(ListProvidersPayload{Providers: p})
}
