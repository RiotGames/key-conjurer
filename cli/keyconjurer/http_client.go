package keyconjurer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	rootcerts "github.com/hashicorp/go-rootcerts"
)

// Creates a httpclient singleton that loads the user's system's CA. Note that
//  this singleton is probably not multi-thread safe.
func getHTTPClientSingleton() (*http.Client, error) {
	certs, err := rootcerts.LoadSystemCAs()
	if err != nil {
		return nil, fmt.Errorf("Could not load System root CA files. Reason: %v", err)
	}

	config := &tls.Config{
		RootCAs: certs,
	}

	tr := &http.Transport{TLSClientConfig: config}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * ClientHttpTimeoutInSeconds,
	}

	return httpClient, nil
}

func createAPIURL(path string) string {
	api := ProdAPI
	if Dev {
		api = DevAPI
	}
	if strings.HasPrefix(path, "/") {
		return fmt.Sprintf("%v%v", api, path)
	} else {
		return fmt.Sprintf("%v/%v", api, path)
	}
}

func doKeyConjurerAPICall(url string, data []byte, responseStruct interface{}) error {
	httpClient, err := getHTTPClientSingleton()
	if err != nil {
		return err
	}

	apiURL := createAPIURL(url)

	req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(data))
	req.Header.Set("content-type", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return errors.New("Error sending http request")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.New("Unable to parse response")
	}

	if res.StatusCode == 200 {
		responseData := &ResponseData{Data: responseStruct}
		if err := json.Unmarshal(body, responseData); err != nil {
			return errors.New("Unable to read json response")
		}

		if responseData.Message != "success" {
			return errors.New(responseData.Message)
		}

		return nil
	}

	errorMessage := ""
	if res.StatusCode >= 400 && res.StatusCode < 500 {
		errorMessage = string(body)
	} else if res.StatusCode >= 500 && res.StatusCode < 600 {
		errorMessage = "Remote host errors"
	} else {
		errorMessage = "Unexpected error occured"
	}
	return errors.New(errorMessage)
}

type UserRequest struct {
	Client             string `json:"client"`
	ClientVersion      string `json:"clientVersion"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	ShouldEncryptCreds bool   `json:"shouldEncryptCreds"`
}

type CredsRequest struct {
	Client         string `json:"client"`
	ClientVersion  string `json:"clientVersion"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	AppID          string `json:"appId"`
	TimeoutInHours uint   `json:"timeoutInHours"`
}

// ResponseData is the standard response structure from the Key Conjurer API
type ResponseData struct {
	Success bool        `json:"Success"`
	Message string      `json:"Message"`
	Data    interface{} `json:"Data"`
}
