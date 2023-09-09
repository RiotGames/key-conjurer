package keyconjurer

import (
	"net/http"
	"strings"
)

func GetBearerToken(r *http.Request) (string, bool) {
	headerValue, ok := r.Header["Authorization"]
	if !ok {
		return "", false
	}

	if len(headerValue) != 1 {
		return "", false
	}

	parts := strings.Split(headerValue[0], " ")
	if len(parts) != 2 {
		return "", false
	}

	return parts[1], parts[0] == "Bearer"
}
