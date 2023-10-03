package main

import (
	"encoding/base64"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/russellhaering/gosaml2/types"
)

type SAMLResponse struct {
	original []byte
	inner    *types.Assertion
}

func (r *SAMLResponse) AddAttribute(name, value string) {
	if r.inner == nil {
		r.inner = &types.Assertion{}
	}

	if r.inner.AttributeStatement == nil {
		r.inner.AttributeStatement = &types.AttributeStatement{}
	}

	val := types.AttributeValue{Type: "xs:string", Value: value}
	r.inner.AttributeStatement.Attributes = append(r.inner.AttributeStatement.Attributes, types.Attribute{
		Name:   name,
		Values: []types.AttributeValue{val},
	})
}

func (r SAMLResponse) GetAttribute(name string) string {
	vals := r.GetAttributeValues(name)
	if len(vals) > 0 {
		return vals[0]
	} else {
		return ""
	}
}

func (r SAMLResponse) GetAttributeValues(name string) []string {
	var vals []string
	for _, attr := range r.inner.AttributeStatement.Attributes {
		if attr.Name == name {
			for _, v := range attr.Values {
				vals = append(vals, v.Value)
			}
		}
	}

	return vals
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

func ParseEncodedResponse(b64ResponseXML string) (*types.Assertion, error) {
	var response types.Assertion
	bytesXML, err := base64.StdEncoding.DecodeString(b64ResponseXML)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(bytesXML, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func ParseBase64EncodedSAMLResponse(xml string) (*SAMLResponse, error) {
	res, err := ParseEncodedResponse(xml)
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
