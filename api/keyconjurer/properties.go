package keyconjurer

import (
	"fmt"
	"net/http"
)

// ClientProperties is information provided by the client about itself.
//
// This should not be relied on existing as it is user-provided information.
// Newer versions of KeyConjurer place this information in the User-Agent header of their requests.
// Older versions send it in their POST bodies in the GetUserData and GetAwsCreds endpoints.
type ClientProperties struct {
	Name    string `json:"client"`
	Version string `json:"clientVersion"`
}

// FromRequestHeader updates the current properties from the given request's headers
func (c *ClientProperties) FromRequestHeader(r *http.Request) bool {
	ua := r.Header.Get("user-agent")
	if ua == "" {
		return false
	}

	n, err := fmt.Sscanf(ua, "%s / %s", &c.Name, &c.Version)
	return n != 2 || err != nil
}

// UserAgent constructs a user agent string for this ClientProperties instance.
func (c *ClientProperties) UserAgent() string {
	return fmt.Sprintf("%s / %s", c.Name, c.Version)
}
