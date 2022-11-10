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
		Timeout: consts.HttpTimeout,
	}
	return Duo{httpClient: duoHTTPClient}
}

func (d Duo) SendMFACode(txSignature, stateToken, callbackURL, apiHostName, code string) (string, error) {
	sid, err := d.getSid(txSignature, stateToken, callbackURL, apiHostName)
	if err != nil {
		return "", err
	}

	d.prepareForPush(sid, txSignature, callbackURL, apiHostName)
	dpr, err := d.sendMfaCode(sid, txSignature, callbackURL, apiHostName, code)
	if err != nil {
		return "", err
	}

	dpr2, err := d.checkMfaStatus(sid, dpr.Response.TxID, apiHostName)
	if err != nil || dpr2.Response.StatusCode == "deny" {
		return "", ErrorDuoMfaNotAllow
	}

	return dpr2.Response.Cookie, nil
}

// SendPush emulates the workflow of the Duo WebAPI and sends the
//
//	requesting user a push to the device set as "phone1"
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
	_, err := d.httpClient.Do(req)
	if err != nil {
		return ErrorDuoCommunication
	}

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

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	duoPromptResponse := duoPromptResponse{}
	if err := dec.Decode(&duoPromptResponse); err != nil {
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

func (d *Duo) sendMfaCode(sid, txSignature, callbackURL, apiHostName, code string) (duoPromptResponse, error) {
	reqURL := fmt.Sprintf("https://%v/frame/prompt?%v",
		apiHostName,
		url.Values{
			"parent": {callbackURL},
			"tx":     {txSignature}}.Encode())

	values := url.Values{
		"sid":      []string{sid},
		"device":   []string{"phone1"},
		"factor":   []string{"Passcode"},
		"passcode": []string{code},
	}

	req, _ := http.NewRequest("POST", reqURL, strings.NewReader(values.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return duoPromptResponse{}, ErrorDuoCommunication
	}

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()

	var dpr duoPromptResponse
	if err := dec.Decode(&dpr); err != nil {
		return duoPromptResponse{}, ErrorJSONMarshalError
	}

	if dpr.Stat != "OK" {
		return duoPromptResponse{}, fmt.Errorf("POST /frame/prompt response body did not have OK stat - Stat=%s", dpr.Stat)
	}

	return dpr, nil
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

	pushResponse := duoPushResponse{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&pushResponse); err != nil {
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

		dec := json.NewDecoder(resp.Body)
		pushResponse = duoPushResponse{}
		if err := dec.Decode(&pushResponse); err != nil {
			return nil, ErrorJSONUnmarshalError
		}
	}

	return &pushResponse, nil
}
