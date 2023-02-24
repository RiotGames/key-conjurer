package keyconjurer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/riotgames/key-conjurer/api/authenticators/duo"
	"github.com/riotgames/key-conjurer/api/authenticators/okta"
	"github.com/riotgames/key-conjurer/api/cloud"
	"github.com/riotgames/key-conjurer/api/consts"
	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/api/settings"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	crypt                   core.Crypto
	cfg                     *settings.Settings
	cloud                   *cloud.Provider
	log                     *logrus.Entry
	authenticationProviders providerMap
}

func NewHandler(cfg *settings.Settings) Handler {
	client, err := cloud.NewProvider(cfg.AwsRegion, cfg.TencentRegion)
	if err != nil {
		panic(err)
	}

	var prov core.CryptoProvider = &core.PassThroughProvider{}
	if cfg.AwsKMSKeyID != "" {
		prov = core.NewKMSProvider(&core.KMSProviderConfig{
			KMSKeyID: cfg.AwsKMSKeyID,
			Session:  session.New(),
		})
	}

	mfa := duo.New()
	return Handler{
		crypt: core.NewCrypto(prov),
		log: newLogger(loggerSettings{
			Level:            logrus.DebugLevel,
			LogstashEndpoint: consts.LogstashEndpoint,
		}),
		cfg:   cfg,
		cloud: client,
		authenticationProviders: providerMap{
			AuthenticationProviderOkta: okta.Must(cfg.OktaHost, cfg.OktaToken, mfa),
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
func (h *Handler) GetUserDataEventHandler(ctx context.Context, req *events.ALBTargetGroupRequest) (*events.ALBTargetGroupResponse, error) {
	if req.HTTPMethod == "OPTIONS" {
		return createAWSResponse(Success, nil)
	}

	log := h.log

	var event GetUserDataEvent
	if err := json.Unmarshal([]byte(req.Body), &event); err != nil {
		log.Errorf("unable to parse incoming JSON: %s", err)
		return ErrorResponse(ErrCodeBadRequest, "unable to parse incoming JSON")
	}

	if err := h.crypt.Decrypt(ctx, &event.Credentials); err != nil {
		log.Errorf("unable to decrypt credentials: %s", err)
		return ErrorResponse(ErrCodeUnableToDecrypt, "unable to decrypt credentials")
	}

	log = h.log.WithFields(logrus.Fields{"username": event.Credentials.Username, "idp": event.AuthenticationProvider})
	provider, ok := h.authenticationProviders.Get(event.AuthenticationProvider)
	if !ok {
		log.Infof("unknown provider %q", provider)
		return ErrorResponse(ErrCodeInvalidProvider, "the provider you supplied is unsupported by this version of KeyConjurer")
	}

	user, err := provider.Authenticate(ctx, event.Credentials)
	if err != nil {
		log.Errorf("failed to authenticate user: %s", err)
		return ErrorResponse(ErrCodeInvalidCredentials, "credentials are incorrect")
	}

	applications, err := provider.ListApplications(ctx, user)
	if err != nil {
		log.Errorf("failed to retrieve applications: %s", err)
		return ErrorResponse(ErrCodeInternalServerError, fmt.Sprintf("failed to retrieve applications: %s", err))
	}

	ciphertext, err := h.crypt.Encrypt(ctx, event.Credentials)
	if err != nil {
		log.Errorf("failed to encrypt credentials: %s", err)
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
	Cloud           int    `json:"Cloud"` // 0:aws,1:tencent
}

// GetTemporaryCredentialEventHandler issues temporary credentials for the current user.
//
// This MUST be backwards compatible with the old version of KeyConjurer for a time.
func (h *Handler) GetTemporaryCredentialEventHandler(ctx context.Context, req *events.ALBTargetGroupRequest) (*events.ALBTargetGroupResponse, error) {
	if req.HTTPMethod == "OPTIONS" {
		return createAWSResponse(Success, nil)
	}

	log := h.log

	var event GetTemporaryCredentialEvent
	if err := json.Unmarshal([]byte(req.Body), &event); err != nil {
		log.Errorf("unable to parse incoming JSON: %s", err)
		return ErrorResponse(ErrCodeBadRequest, "unable to parse incoming JSON")
	}

	if err := event.Validate(); err != nil {
		log.Infof("bad request: %s", err.Error())
		return ErrorResponse(ErrCodeBadRequest, err.Error())
	}

	if err := h.crypt.Decrypt(ctx, &event.Credentials); err != nil {
		log.Errorf("unable to decrypt credentials: %s", err)
		return ErrorResponse(ErrCodeUnableToDecrypt, "unable to decrypt credentials")
	}

	log = h.log.WithFields(logrus.Fields{"username": event.Credentials.Username, "idp": event.AuthenticationProvider, "account_id": event.AppID})
	provider, ok := h.authenticationProviders.Get(event.AuthenticationProvider)
	if !ok {
		log.Infof("unknown provider %q", provider)
		return ErrorResponse(ErrCodeInvalidProvider, "invalid provider")
	}

	if _, err := provider.Authenticate(ctx, event.Credentials); err != nil {
		log.Errorf("failed to authenticate user: %s", err)
		return ErrorResponse(ErrCodeInvalidCredentials, "credentials are incorrect")
	}

	response, err := provider.GenerateSAMLAssertion(ctx, event.Credentials, event.AppID)
	if err != nil {
		msg := fmt.Sprintf("unable to generate SAML assertion: %s", err)
		log.Errorf(msg)
		return ErrorResponse(getErrorCode(err), msg)
	}

	cloudFlag, sts, err := h.cloud.GetTemporaryCredentialsForUser(ctx, event.RoleName, response, int(event.TimeoutInHours))
	if err != nil {
		var errRoleNotFound cloud.ErrRoleNotFound
		if errors.As(err, &errRoleNotFound) {
			log.Infof("role %q either does not exist or the user is not entitled to it", event.RoleName)
			return ErrorResponse(ErrCodeBadRequest, errRoleNotFound.Error())
		}

		log.Errorf("failed to generate temporary session credentials: %s", err.Error())
		return ErrorResponse(ErrCodeInternalServerError, err.Error())
	}

	return DataResponse(GetTemporaryCredentialsPayload{
		AccountID:       event.AppID,
		AccessKeyID:     *sts.AccessKeyID,
		SecretAccessKey: *sts.SecretAccessKey,
		SessionToken:    *sts.SessionToken,
		Expiration:      sts.Expiration,
		Cloud:           cloudFlag,
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
func (h *Handler) ListProvidersHandler(ctx context.Context) (*events.ALBTargetGroupResponse, error) {
	var p []Provider
	for key := range h.authenticationProviders {
		p = append(p, Provider{ID: key})
	}

	return DataResponse(ListProvidersPayload{Providers: p})
}

// getErrorCode translates an error to an ErrorCode.
func getErrorCode(err error) ErrorCode {
	switch {
	case errors.Is(err, core.ErrInternalError):
		return ErrCodeInternalServerError
	case errors.Is(err, core.ErrBadRequest):
		return ErrCodeBadRequest
	case errors.Is(err, core.ErrAuthenticationFailed):
		return ErrCodeInvalidCredentials
	case errors.Is(err, core.ErrAccessDenied):
		return ErrCodeInvalidCredentials
	default:
		return ErrCodeUnspecified
	}
}
