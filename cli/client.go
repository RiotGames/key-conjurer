package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"

	rootcerts "github.com/hashicorp/go-rootcerts"
)

type Client struct {
	baseURL url.URL
	http    *http.Client
}

func NewClient(baseURL url.URL) (Client, error) {
	certs, err := rootcerts.LoadSystemCAs()
	if err != nil {
		return Client{}, fmt.Errorf("could not load System root CA files. Reason: %v", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certs,
			},
		},
		Timeout: time.Second * time.Duration(clientHttpTimeoutSeconds),
	}

	return Client{http: httpClient, baseURL: baseURL}, nil
}

func getBinaryName() string {
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			return LinuxArm64BinaryName
		}

		return LinuxAmd64BinaryName
	case "windows":
		return WindowsBinaryName
	default:
		if runtime.GOARCH == "arm64" {
			return DarwinArm64BinaryName
		}
		return DarwinAmd64BinaryName
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
