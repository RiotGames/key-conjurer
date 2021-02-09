package keyconjurer

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
)

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
func GetLatestBinary() ([]byte, error) {
	httpClient, err := createHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("could not get HTTP client: %w", err)
	}

	binaryURL := fmt.Sprintf("%s/%s", DownloadURL, getBinaryName())
	req, _ := http.NewRequest(http.MethodGet, binaryURL, nil)
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get binary upgrade: %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("Unable to parse response")
	}
	return body, nil
}
