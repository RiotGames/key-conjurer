package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/RobotsAndPencils/go-saml"
)

type SAMLResponse struct {
	original []byte
	inner    *saml.Response
}

func (r *SAMLResponse) AddAttribute(name, value string) {
	if r.inner == nil {
		r.inner = &saml.Response{}
	}

	r.inner.AddAttribute(name, value)
}

func (r SAMLResponse) GetAttribute(name string) string {
	return r.inner.GetAttribute(name)
}

func (r SAMLResponse) GetAttributeValues(name string) []string {
	return r.inner.GetAttributeValues(name)
}

type RoleProviderPair struct {
	RoleARN     string
	ProviderARN string
}

func ListSAMLRoles(response *SAMLResponse) []string {
	if response == nil {
		return nil
	}

	roleURL := "https://aws.amazon.com/SAML/Attributes/Role"
	roleSubstr := "role/"
	if response.GetAttribute(roleURL) == "" {
		roleURL = "https://cloud.tencent.com/SAML/Attributes/Role"
		roleSubstr = "roleName/"
	}

	var names []string
	for _, v := range response.GetAttributeValues(roleURL) {
		p := getARN(v)
		idx := strings.Index(p.RoleARN, roleSubstr)
		parts := strings.Split(p.RoleARN[idx:], "/")
		names = append(names, parts[1])
	}

	return names
}

func FindRoleInSAML(roleName string, response *SAMLResponse) (RoleProviderPair, bool) {
	if response == nil {
		return RoleProviderPair{}, false
	}

	roleURL := "https://aws.amazon.com/SAML/Attributes/Role"
	roleSubstr := "role/"
	attrs := response.GetAttributeValues(roleURL)
	if len(attrs) == 0 {
		attrs = response.GetAttributeValues("https://cloud.tencent.com/SAML/Attributes/Role")
		roleSubstr = "roleName/"
	}

	if len(attrs) == 0 {
		// The SAML assertoin contains no known roles for AWS or Tencent.
		return RoleProviderPair{}, false
	}

	var pairs []RoleProviderPair
	for _, v := range response.GetAttributeValues(roleURL) {
		pairs = append(pairs, getARN(v))
	}

	if len(pairs) == 0 {
		return RoleProviderPair{}, false
	}

	var pair RoleProviderPair
	for _, p := range pairs {
		idx := strings.Index(p.RoleARN, roleSubstr)
		parts := strings.Split(p.RoleARN[idx:], "/")
		if strings.EqualFold(parts[1], roleName) {
			pair = p
		}
	}

	if pair.RoleARN == "" {
		return RoleProviderPair{}, false
	}

	return pair, true
}

func getARN(value string) RoleProviderPair {
	p := RoleProviderPair{}
	roles := strings.Split(value, ",")
	if len(roles) >= 2 {
		if strings.Contains(roles[0], "saml-provider/") {
			p.ProviderARN = roles[0]
			p.RoleARN = roles[1]
		} else {
			p.ProviderARN = roles[1]
			p.RoleARN = roles[0]
		}
	}
	return p
}

func ParseBase64EncodedSAMLResponse(xml string) (*SAMLResponse, error) {
	res, err := saml.ParseEncodedResponse(xml)
	if err != nil {
		return nil, nil
	}
	return &SAMLResponse{original: []byte(xml), inner: res}, nil
}

type SAMLCallbackHandler struct {
	AssertionChannel chan []byte
}

func (h SAMLCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Handle panics gracefully
	assertionBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Failed to read request body: %s", err)
	}

	// A correctly formed body will be a url-encoded param response, with the SAMLResponse being base64 encoded.
	v, err := url.ParseQuery(string(assertionBytes))
	if err != nil {
		log.Fatalf("Incorrectly formatted request body: %s", err)
	}
	r.Body.Close()

	// TODO: This can panic if the body doesn't contain SAMLResponse
	h.AssertionChannel <- []byte(v["SAMLResponse"][0])
}
