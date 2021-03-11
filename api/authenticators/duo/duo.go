package duo

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/riotgames/key-conjurer/api/consts"
)

var (
	ErrorDuoCommunication   = errors.New("unable to communicate with Duo")
	ErrorDuoMfaNotAllow     = errors.New("MFA was not allowed")
	ErrorDuoArgsError       = errors.New("there was an error parsing arguments for Duo push request")
	ErrorDuoPushError       = errors.New("there was an error sending a Duo push request")
	ErrorCannotFindSid      = errors.New("Cannot find sid")
	ErrorHTTPBodyError      = errors.New("Unable to read http response body")
	ErrorJSONMarshalError   = errors.New("Unable to marshal json")
	ErrorJSONUnmarshalError = errors.New("Unable to unmarshal json")
)

// Duo scripts the Duo Web API interaction
type Duo struct {
	httpClient *http.Client
}

type duoPromptResponse struct {
	Stat     string `json:"stat"`
	Response struct {
		TxID string `json:"txid"`
	} `json:"response"`
}

type pushResponse struct {
	StatusCode string `json:"status_code"`
	Parent     string `json:"parent"`
	Result     string `json:"result"`
	Cookie     string `json:"cookie"`
	ResultURL  string `json:"result_url"`
}

type duoPushResponse struct {
	Stat     string       `json:"stat"`
	Response pushResponse `json:"response"`
}

// New returns a new Duo client that uses the provided logger
func New() Duo {
	duoHTTPClient := &http.Client{
		Timeout: time.Second * consts.HttpTimeoutInSeconds,
	}
	return Duo{httpClient: duoHTTPClient}
}

// SendPush emulates the workflow of the Duo WebAPI and sends the
//  requesting user a push to the device set as "phone1"
func (d *Duo) SendPush(txSignature, stateToken, callbackURL, apiHostName string) (string, error) {
	sid, err := d.getSid(txSignature, stateToken, callbackURL, apiHostName)
	if err != nil {
		return "", err
	}
	d.prepareForPush(sid, txSignature, callbackURL, apiHostName)
	txid, err := d.sendMfaPush(sid, txSignature, callbackURL, apiHostName)
	if err != nil {
		return "", err
	}
	mfaStatus, err := d.checkMfaStatus(sid, txid, apiHostName)
	if err != nil {
		return "", err
	}
	if strings.ToLower(mfaStatus.Stat) != "ok" {
		return "", ErrorDuoMfaNotAllow
	}
	return mfaStatus.Response.Cookie, nil
}

func (d *Duo) getSid(txSignature, stateToken, callbackURL, apiHostName string) (string, error) {
	reqURL := fmt.Sprintf("https://%v/frame/web/v1/auth?%v",
		apiHostName,
		url.Values{
			"parent": {callbackURL},
			"tx":     {txSignature}}.Encode())

	req, _ := http.NewRequest("POST", reqURL, strings.NewReader(url.Values{"parent": {callbackURL}}.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", ErrorDuoCommunication
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", ErrorHTTPBodyError
	}

	// View duo_test.go "Test_Duo_getSid" to see the expected cases that this
	//  regex will match
	data := regexp.MustCompile("<input.*name=['\"]?sid['\"]?.*?value=['\"]?(.*?)['\"]?>").FindSubmatch(body)
	if len(data) != 2 {
		return "", ErrorCannotFindSid
	}

	sid := html.UnescapeString(string(data[1]))
	return sid, nil
}

func (d *Duo) prepareForPush(sid, txSignature, callbackURL, apiHostName string) error {
	const certsURL = "https://certs-duo1.duosecurity.com/frame/client_cert"

	reqURL := fmt.Sprintf("https://%v/frame/web/v1/auth?%v",
		apiHostName,
		url.Values{
			"parent": {callbackURL},
			"tx":     {txSignature}}.Encode())
	req, _ := http.NewRequest("POST", reqURL, strings.NewReader(url.Values{
		"sid":       []string{sid},
		"certs_url": []string{certsURL}}.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return ErrorDuoCommunication
	}

	defer resp.Body.Close()
	return nil
}

func (d *Duo) sendMfaPush(sid, txSignature, callbackURL, apiHostName string) (string, error) {
	reqURL := fmt.Sprintf("https://%v/frame/prompt?%v",
		apiHostName,
		url.Values{
			"parent": {callbackURL},
			"tx":     {txSignature}}.Encode())
	req, _ := http.NewRequest("POST", reqURL, strings.NewReader(url.Values{
		"sid":    []string{sid},
		"device": []string{"phone1"},
		"factor": []string{"Duo Push"}}.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", ErrorDuoCommunication
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", ErrorHTTPBodyError
	}

	duoPromptResponse := &duoPromptResponse{}
	if err := json.Unmarshal(body, duoPromptResponse); err != nil {
		return "", ErrorJSONMarshalError
	}

	// Checking MFA status does not block on the first attempt
	//  to check the status after pushing.  The response just says
	//  that a push was sent.
	mfaResponse, err := d.checkMfaStatus(sid, duoPromptResponse.Response.TxID, apiHostName)
	if err != nil {
		return "", ErrorDuoPushError
	}

	if strings.ToLower(mfaResponse.Response.StatusCode) != "pushed" {
		return "", ErrorDuoPushError
	}

	return duoPromptResponse.Response.TxID, nil
}

func (d *Duo) checkMfaStatus(sid, txid, apiHostName string) (*duoPushResponse, error) {
	reqURL := fmt.Sprintf("https://%v/frame/status", apiHostName)
	req, _ := http.NewRequest("POST", reqURL, strings.NewReader(url.Values{
		"sid":  []string{sid},
		"txid": []string{txid}}.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, ErrorDuoCommunication
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorHTTPBodyError
	}

	pushResponse := &duoPushResponse{}
	if err := json.Unmarshal(body, pushResponse); err != nil {
		return nil, ErrorJSONUnmarshalError
	}

	// Duo changed their workflow so actual needed response is one more
	//  request after verifying the push.  This handles the complete
	//  flow to retrieve the proper response.
	if pushResponse.Response.ResultURL != "" {
		reqURL := fmt.Sprintf("https://%v%v", apiHostName, pushResponse.Response.ResultURL)
		req, _ := http.NewRequest("POST", reqURL, strings.NewReader(url.Values{
			"sid": []string{sid}}.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := d.httpClient.Do(req)
		if err != nil {
			return nil, ErrorDuoCommunication
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, ErrorHTTPBodyError
		}
		pushResponse = &duoPushResponse{}
		if err := json.Unmarshal(body, pushResponse); err != nil {
			return nil, ErrorJSONUnmarshalError
		}
	}

	return pushResponse, nil
}
