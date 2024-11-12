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

	w.Headers["Content-Type"] = "application/json"
	w.Body = string(buf)
}

func ServeJSONError(w *events.ALBTargetGroupResponse, statusCode int, msg string) {
	var jsonError struct {
		Message string `json:"error"`
	}

	jsonError.Message = msg
	w.StatusCode = statusCode
	ServeJSON(w, jsonError)
}
