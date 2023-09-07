package httputil

import (
	"encoding/json"
	"net/http"
)

func ServeJSON[T any](w http.ResponseWriter, data T) {
	buf, err := json.Marshal(data)
	if err != nil {
		// Nothing to be done here
		return
	}

	// Nothing we can do to respond to the error message here either
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

type JSONError struct {
	Message string `json:"error"`
}

func ServeJSONError(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	ServeJSON[JSONError](w, JSONError{msg})
}
