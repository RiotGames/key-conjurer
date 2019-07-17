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
		return "keyconjurer-linux"
	case "windows":
		return "keyconjurer-windows.exe"
	default:
		return "keyconjurer-darwin"
	}
}

// GetLatestBinary downloads the latest keyconjurer binary from the web.
func GetLatestBinary() ([]byte, error) {
	httpClient, err := getHTTPClientSingleton()
	if err != nil {
		Logger.Warnln("Error getting http client")
		return nil, err
	}

	binaryURL := fmt.Sprintf("%s/%s", DownloadURL, getBinaryName())
	req, _ := http.NewRequest(http.MethodGet, binaryURL, nil)
	res, err := httpClient.Do(req)
	if err != nil {
		Logger.Errorf("Could not get binary upgrade. Reason: %v", err)
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Logger.Errorln(err)
		return nil, errors.New("Unable to parse response")
	}
	return body, nil
}
