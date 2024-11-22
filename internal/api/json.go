package api

import (
	"encoding/json"

	"log/slog"

	"github.com/aws/aws-lambda-go/events"
)

func ServeJSON[T any](w *events.ALBTargetGroupResponse, data T) {
	buf, err := json.Marshal(data)
	if err != nil {
		// Nothing to be done here
		slog.Error("could not marshal JSON: %s", "error", err)
		return
	}

	if w.Headers == nil {
		w.Headers = make(map[string]string)
	}

	w.Headers["Content-Type"] = "application/json"
	w.Body = string(buf)
}

type JSONError struct {
	Message string `json:"error"`
}

func ServeJSONError(w *events.ALBTargetGroupResponse, statusCode int, msg string) {
	w.StatusCode = statusCode
	ServeJSON(w, JSONError{Message: msg})
}
