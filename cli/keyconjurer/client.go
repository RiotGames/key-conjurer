package keyconjurer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"time"

	rootcerts "github.com/hashicorp/go-rootcerts"
	"github.com/riotgames/key-conjurer/api/core"
	api "github.com/riotgames/key-conjurer/api/keyconjurer"
)

// client and version are injected at compile time, refer to consts.go
var props api.ClientProperties = api.ClientProperties{
	Name:    ClientName,
	Version: Version,
}

var (
	errUnspecifiedServerError = errors.New("unspecified server error")
	errInvalidJSONResponse    = errors.New("unable to parse JSON from the server")
)

func createAPIURL(hostname, path string) string {
	if strings.HasPrefix(path, "/") {
		return fmt.Sprintf("%v%v", hostname, path)
	}

	return fmt.Sprintf("%v/%v", hostname, path)
}

// Client is the struct through which all KeyConjurer operations stem.
type Client struct {
	hostname string
	http     *http.Client
}

// New creates a new client with the given hostname.
func New(hostname string) (Client, error) {
	certs, err := rootcerts.LoadSystemCAs()
	if err != nil {
		return Client{}, fmt.Errorf("Could not load System root CA files. Reason: %v", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		},
		Timeout: time.Second * ClientHttpTimeoutInSeconds,
	}

	return Client{http: httpClient, hostname: hostname}, nil
}

func (c *Client) do(ctx context.Context, method, url string, data []byte, responseStruct interface{}) error {
	apiURL := createAPIURL(c.hostname, url)
	req, err := http.NewRequestWithContext(ctx, method, apiURL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	// We use the User Agent header to indicate client versions in newer versions of Key Conjurer because you cannot send bodies with GET
	req.Header.Set("user-agent", props.UserAgent())
	req.Header.Set("content-type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("error sending http request: %w", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to parse response: %w", err)
	}

	// With our AWS Lambda setup, the server will always return HTTP 200 unless there was an internal error on the server's end.
	if res.StatusCode >= 500 {
		return errUnspecifiedServerError
	}

	var response api.Response
	if err := json.Unmarshal(body, &response); err != nil {
		return errInvalidJSONResponse
	}

	if !response.Success {
		var responseError error
		err := response.GetError(&responseError)
		if err != nil {
			return err
		}

		return responseError
	}

	return response.GetPayload(&responseStruct)
}

type GetCredentialsOptions struct {
	Credentials            core.Credentials
	ApplicationID          string
	TimeoutInHours         uint8
	RoleName               string
	AuthenticationProvider api.AuthenticationProviderName
}

// GetCredentials requests a set of temporary credentials for the requested AWS account and returns them.
func (c *Client) GetCredentials(ctx context.Context, opts *GetCredentialsOptions) (*AWSCredentials, error) {
	request := api.GetTemporaryCredentialEvent{
		Credentials:            opts.Credentials,
		AppID:                  opts.ApplicationID,
		TimeoutInHours:         opts.TimeoutInHours,
		RoleName:               opts.RoleName,
		AuthenticationProvider: opts.AuthenticationProvider,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	var response api.GetTemporaryCredentialsPayload
	if err := c.do(ctx, "POST", "/get_aws_creds", data, &response); err != nil {
		return nil, fmt.Errorf("failed to generate temporary session token: %s", err.Error())
	}

	aws := AWSCredentials{
		AccountID:       response.AccountID,
		AccessKeyID:     response.AccessKeyID,
		SecretAccessKey: response.SecretAccessKey,
		SessionToken:    response.SessionToken,
		Expiration:      response.Expiration,
	}

	return &aws, nil
}

// GetUserData returns data on the user stored in the API.
func (c *Client) GetUserData(ctx context.Context, creds core.Credentials) (api.GetUserDataPayload, error) {
	// TODO: This is broken; the server does not return a UserData struct but an api.GetUserDataPayload struct
	var ud UserData
	request := api.GetUserDataEvent{
		Credentials:            creds,
		AuthenticationProvider: api.AuthenticationProviderOkta,
	}

	b, err := json.Marshal(request)
	if err != nil {
		return api.GetUserDataPayload{}, err
	}

	var data api.GetUserDataPayload
	return data, c.do(ctx, "POST", "/get_user_data", b, &ud)
}

type ListAccountsOptions struct {
	AuthenticationProvider api.AuthenticationProviderName
	Credentials            core.Credentials
}

// ListAccounts lists the accounts the user is entitled to access.
func (c *Client) ListAccounts(ctx context.Context, opts *ListAccountsOptions) ([]core.Application, error) {
	// HACK: We can re-use the GetUserData endpoint as it returns the applications the user is entitled to view.
	payload := api.GetUserDataEvent{
		Credentials:            opts.Credentials,
		AuthenticationProvider: opts.AuthenticationProvider,
	}

	if opts.AuthenticationProvider == "" {
		payload.AuthenticationProvider = api.AuthenticationProviderOkta
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var response api.GetUserDataPayload
	if err := c.do(ctx, "POST", "/get_user_data", data, &response); err != nil {
		return nil, err
	}

	return response.Apps, nil
}

type ListRolesOptions struct {
	AuthenticationProvider api.AuthenticationProviderName
	Credentials            core.Credentials
	AccountID              string
}

// ListRoles lists the roles the user with the given credentials may assume for the given account.
//
// If AuthenticationProvider is not specified in ListRoleOptions, it will default to Okta.
func (c *Client) ListRoles(ctx context.Context, opts *ListRolesOptions) ([]core.Role, error) {
	payload := api.ListRolesEvent{
		Credentials: opts.Credentials,
		Provider:    opts.AuthenticationProvider,
	}

	if opts.AuthenticationProvider == "" {
		payload.Provider = api.AuthenticationProviderOkta
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var result api.ListRolesPayload
	return result.Roles, c.do(ctx, "GET", "/roles", b, &result)
}

type ListProvidersOptions struct{}

// ListProviders lists the authentication providers that the user may use to authenticate with.
func (c *Client) ListProviders(ctx context.Context, opts *ListProvidersOptions) ([]api.Provider, error) {
	payload := api.ListProvidersEvent{}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var result api.ListProvidersPayload
	return result.Providers, c.do(ctx, "GET", "/authentication_providers", b, &result)
}

func getBinaryName() string {
	switch runtime.GOOS {
	case "linux":
		return LinuxBinaryName
	case "windows":
		return WindowsBinaryName
	default:
		return DarwinBinaryName
	}
}

// GetLatestBinary downloads the latest keyconjurer binary from the web.
func (c *Client) GetLatestBinary(ctx context.Context) ([]byte, error) {
	binaryURL := fmt.Sprintf("%s/%s", DownloadURL, getBinaryName())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, binaryURL, nil)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get binary upgrade: %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("Unable to parse response")
	}

	return body, nil
}
