package vault

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	hashivault "github.com/hashicorp/vault/api"
)

type IAM struct {
	client *hashivault.Client
}

type IAMLoginOptions struct {
	Role      string
	MountPath string
}

func (i *IAM) Login(options IAMLoginOptions) (*hashivault.Secret, error) {
	authSecret, err := i.iamLogin(options)

	if err != nil {
		return nil, err
	}

	if authSecret.Auth == nil {
		return nil, errors.New("Vault IAM Auth returned nil")
	}

	i.client.SetToken(authSecret.Auth.ClientToken)
	return authSecret, nil
}

// This is based on https://github.com/daveadams/onthelambda/blob/7eb4dc8a8cb58b8a17ba19ad5c68bb4affcbdd22/onthelambda.go#L118-L158
//  Credit to @daveadams
func (i *IAM) iamLogin(options IAMLoginOptions) (*hashivault.Secret, error) {
	stsClient := sts.New(session.Must(session.NewSession()))

	req, _ := stsClient.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	if err := req.Sign(); err != nil {
		return nil, err
	}

	headers, err := json.Marshal(req.HTTPRequest.Header)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(req.HTTPRequest.Body)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"iam_http_request_method": req.HTTPRequest.Method,
		"iam_request_url":         base64.StdEncoding.EncodeToString([]byte(req.HTTPRequest.URL.String())),
		"iam_request_headers":     base64.StdEncoding.EncodeToString(headers),
		"iam_request_body":        base64.StdEncoding.EncodeToString(body),
		"role":                    options.Role}

	authPath := "auth/aws/login"
	if options.MountPath != "" {
		authPath = "auth/" + strings.Trim(options.MountPath, "/") + "/login"
	}

	resp, err := i.client.Logical().Write(authPath, data)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("Got no response from the aws authentication provider")
	}

	return resp, nil
}
