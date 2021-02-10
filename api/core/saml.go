package core

import (
	"github.com/RobotsAndPencils/go-saml"
)

// SAMLResponse contains a raw SAML Response from an IdP.
// This is used to provide access to the original, signed SAML response from the IdP as parsing it into XML and then attempting to encode it again loses this information.
type SAMLResponse struct {
	saml.Response

	// original is the original base64 encoded SAML response from the IdP, including signing information
	original string
}

// GetBase64Encoded returns the base64 encoded SAML response from the IdP.
func (s *SAMLResponse) GetBase64Encoded() *string {
	return &s.original
}

// ParseEncodedResponse parses the base64-encoded SAML assertion provided and returns a SAMLResponse object
func ParseEncodedResponse(b64EncodedXML string) (*SAMLResponse, error) {
	resp, err := saml.ParseEncodedResponse(b64EncodedXML)
	if err != nil {
		return nil, err
	}

	return &SAMLResponse{Response: *resp, original: b64EncodedXML}, nil
}
