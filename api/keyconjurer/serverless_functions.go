package keyconjurer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/riotgames/key-conjurer/api/consts"
	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/api/settings"
	"github.com/riotgames/key-conjurer/internal"
	"github.com/riotgames/key-conjurer/pkg/httputil"
	"github.com/riotgames/key-conjurer/providers"
	kcokta "github.com/riotgames/key-conjurer/providers/okta"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slog"
)

type Handler struct {
	crypt core.Crypto
	cfg   *settings.Settings
	cloud *internal.Provider
	log   *logrus.Entry
}

func NewHandler(cfg *settings.Settings) Handler {
	client, err := internal.NewProvider(cfg.AwsRegion, cfg.TencentRegion)
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

	ok, err := kcokta.New(cfg.OktaHost, cfg.OktaToken)
	if err != nil {
		panic(err)
	}

	providers.Register(providers.Okta, &ok)

	return Handler{
		crypt: core.NewCrypto(prov),
		log: newLogger(loggerSettings{
			Level:            logrus.DebugLevel,
			LogstashEndpoint: consts.LogstashEndpoint,
		}),
		cfg:   cfg,
		cloud: client,
	}
}

type GetUserDataEvent struct {
	core.Credentials
	// AuthenticationProvider is the authentication provider that should be used when logging in.
	AuthenticationProvider string `json:"authentication_provider"`
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

	log = h.log.WithFields(logrus.Fields{
		"username":   event.Credentials.Username,
		"idp":        event.AuthenticationProvider,
		"user_agent": req.Headers["user-agent"],
	})

	provider, ok := providers.Get(event.AuthenticationProvider)
	if !ok {
		log.Infof("unknown provider %q", provider)
		return ErrorResponse(ErrCodeInvalidProvider, "the provider you supplied is unsupported by this version of KeyConjurer")
	}

	user, err := provider.Authenticate(ctx, providers.Credentials(event.Credentials))
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

	log.Info("GetUserDataEventHandler success")
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
	AuthenticationProvider string `json:"authentication_provider"`
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

	if e.AuthenticationProvider == providers.OneLogin {
		// We don't use role names in OneLogin
		e.RoleName = ""
	}

	if e.AuthenticationProvider == providers.Okta && e.RoleName == "" {
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

	log = h.log.WithFields(logrus.Fields{
		"username":   event.Credentials.Username,
		"idp":        event.AuthenticationProvider,
		"user_agent": req.Headers["user-agent"],
		"account_id": event.AppID,
	})

	provider, ok := providers.Get(event.AuthenticationProvider)
	if !ok {
		log.Infof("unknown provider %q", provider)
		return ErrorResponse(ErrCodeInvalidProvider, "invalid provider")
	}

	response, err := provider.GenerateSAMLAssertion(ctx, providers.Credentials(event.Credentials), event.AppID)
	if err != nil {
		log.Errorf("Unable to authenticate user. The credentials may be incorrect, or something may have gone wrong internally. Reason: %s", err)
		return ErrorResponse(getErrorCode(err), "Unable to authenticate. Your credentials may be incorrect. Please contact your system administrators if you're unsure of what to do.")
	}

	sts, err := h.cloud.GetTemporaryCredentialsForUser(ctx, event.RoleName, response, int(event.TimeoutInHours))
	if err != nil {
		var errRoleNotFound internal.ErrRoleNotFound
		if errors.As(err, &errRoleNotFound) {
			log.Infof("role %q either does not exist or the user is not entitled to it", event.RoleName)
			return ErrorResponse(ErrCodeBadRequest, errRoleNotFound.Error())
		}

		log.Errorf("failed to generate temporary session credentials: %s", err.Error())
		return ErrorResponse(ErrCodeInternalServerError, err.Error())
	}

	log.Info("GetTemporaryCredentialEventHandler success")
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
func (h *Handler) ListProvidersHandler(ctx context.Context) (*events.ALBTargetGroupResponse, error) {
	var p []Provider

	providers.ForEach(func(name string, _ providers.Provider) {
		p = append(p, Provider{ID: name})
	})

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
			httputil.ServeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		info, err := okta.GetUserInfo(ctx, idToken)
		if err != nil {
			requestAttrs = append(requestAttrs, slog.String("error", err.Error()))
			slog.Error("okta rejected id token", requestAttrs...)
			httputil.ServeJSONError(w, http.StatusForbidden, "unauthorized")
			return
		}

		requestAttrs = append(requestAttrs, slog.String("username", info.PreferredUsername))
		applications, err := okta.ListApplicationsForUser(ctx, info.PreferredUsername)
		if err != nil {
			requestAttrs = append(requestAttrs, slog.String("error", err.Error()))
			slog.Error("failed to fetch applications", requestAttrs...)
			httputil.ServeJSONError(w, http.StatusBadGateway, "upstream error")
			return
		}

		var accounts []core.Application
		for i, app := range applications {
			if app.AppName == "amazon_aws" || strings.Contains(app.AppName, "tencent") {
				accounts[i] = core.Application{
					ID:   app.Id,
					Name: app.Label,
				}
			}
		}

		requestAttrs = append(requestAttrs, slog.Int("application_count", len(accounts)))
		slog.Info("served applications", requestAttrs...)
		httputil.ServeJSON(w, accounts)
	})
}
