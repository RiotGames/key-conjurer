package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	rootcerts "github.com/hashicorp/go-rootcerts"
	"github.com/riotgames/key-conjurer/api/core"
	"github.com/riotgames/key-conjurer/api/keyconjurer"
)

// client and version are injected at compile time, refer to consts.go
var props keyconjurer.ClientProperties = keyconjurer.ClientProperties{
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

// NewClient creates a new client with the given hostname.
func NewClient(hostname string) (Client, error) {
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

func (c *Client) do(ctx context.Context, url string, r io.Reader, responseStruct interface{}) error {
	apiURL := createAPIURL(c.hostname, url)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, r)
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

	// With our AWS Lambda setup, the server will always return HTTP 200 unless there was an internal error on the server's end.
	if res.StatusCode >= 500 {
		return errUnspecifiedServerError
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()
	var response keyconjurer.Response
	if err := dec.Decode(&response); err != nil {
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
	AuthenticationProvider keyconjurer.AuthenticationProviderName
}

func (c *Client) encodeJSON(data interface{}) (bytes.Buffer, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	// This will prevent certain characters causing problems if they are present in passwords
	enc.SetEscapeHTML(false)
	return b, enc.Encode(data)
}

// GetCredentials requests a set of temporary credentials for the requested AWS account and returns them.
func (c *Client) GetCredentials(ctx context.Context, opts *GetCredentialsOptions) (*AWSCredentials, error) {
	request := keyconjurer.GetTemporaryCredentialEvent{
		Credentials:            opts.Credentials,
		AppID:                  opts.ApplicationID,
		TimeoutInHours:         opts.TimeoutInHours,
		RoleName:               opts.RoleName,
		AuthenticationProvider: opts.AuthenticationProvider,
	}

	buf, err := c.encodeJSON(request)
	if err != nil {
		return nil, err
	}

	var response keyconjurer.GetTemporaryCredentialsPayload
	if err := c.do(ctx, "/get_aws_creds", &buf, &response); err != nil {
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

type GetUserDataOptions struct {
	Credentials            core.Credentials
	AuthenticationProvider keyconjurer.AuthenticationProviderName
}

// GetUserData returns data on the user stored in the API.
func (c *Client) GetUserData(ctx context.Context, opts *GetUserDataOptions) (keyconjurer.GetUserDataPayload, error) {
	request := keyconjurer.GetUserDataEvent{
		Credentials:            opts.Credentials,
		AuthenticationProvider: opts.AuthenticationProvider,
	}

	buf, err := c.encodeJSON(request)
	if err != nil {
		return keyconjurer.GetUserDataPayload{}, err
	}

	var data keyconjurer.GetUserDataPayload
	return data, c.do(ctx, "/get_user_data", &buf, &data)
}

type ListAccountsOptions struct {
	AuthenticationProvider keyconjurer.AuthenticationProviderName
	Credentials            core.Credentials
}

// ListAccounts lists the accounts the user is entitled to access.
func (c *Client) ListAccounts(ctx context.Context, opts *ListAccountsOptions) ([]core.Application, error) {
	// HACK: We can re-use the GetUserData endpoint as it returns the applications the user is entitled to view.
	payload := keyconjurer.GetUserDataEvent{
		Credentials:            opts.Credentials,
		AuthenticationProvider: opts.AuthenticationProvider,
	}

	if opts.AuthenticationProvider == "" {
		payload.AuthenticationProvider = keyconjurer.AuthenticationProviderOkta
	}

	buf, err := c.encodeJSON(payload)
	if err != nil {
		return nil, err
	}

	var response keyconjurer.GetUserDataPayload
	if err := c.do(ctx, "/get_user_data", &buf, &response); err != nil {
		return nil, err
	}

	return response.Apps, nil
}

type ListProvidersOptions struct{}

// ListProviders lists the authentication providers that the user may use to authenticate with.
func (c *Client) ListProviders(ctx context.Context, opts *ListProvidersOptions) ([]keyconjurer.Provider, error) {
	buf, err := c.encodeJSON(keyconjurer.ListProvidersEvent{})
	if err != nil {
		return nil, err
	}

	var result keyconjurer.ListProvidersPayload
	return result.Providers, c.do(ctx, "/list_providers", &buf, &result)
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
func (c *Client) DownloadLatestBinary(ctx context.Context, w io.Writer) error {
	binaryURL := fmt.Sprintf("%s/%s", DownloadURL, getBinaryName())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, binaryURL, nil)
	if err != nil {
		return fmt.Errorf("could not upgrade: %w", err)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("could not upgrade: %w", err)
	}

	if res.StatusCode != 200 {
		return errors.New("could not upgrade: response did not indicate success - are you being blocked by the server?")
	}

	_, err = io.Copy(w, res.Body)
	return err
}
