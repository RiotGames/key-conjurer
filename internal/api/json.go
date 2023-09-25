package api

import (
	"encoding/json"
	"net/http"

	"golang.org/x/exp/slog"
)

func ServeJSON[T any](w http.ResponseWriter, data T) {
	buf, err := json.Marshal(data)
	if err != nil {
		// Nothing to be done here
		slog.Error("could not marshal JSON: %s", err)
		return
	}

	// Nothing we can do to respond to the error message here either
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(buf)
	if err != nil {
		slog.Error("could not write JSON to the client: %s", err)
	}
}

type JSONError struct {
	Message string `json:"error"`
}

func ServeJSONError(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	ServeJSON[JSONError](w, JSONError{msg})
}
