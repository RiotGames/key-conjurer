package keyconjurer

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Response is the generic structure of the lambda responses.
//
// The unmarshalling of this structure is particularly poorly handled due to backwards compatibility concerns.
type Response struct {
	Success bool
	// DEPRECATED: Use ErrorData.Message in the Data field if you intend to communicate error messages to the user.
	Message string
	// Data is the data that will be shipped to the user.
	// Because it is not possible to UnmarshalJSON to interface{}, you must use the GetPayload() or GetError() functions instead to interact with this when unmarshalling from JSON.
	// This is a bad pattern but it's a quick fix.
	Data interface{}
}

func (r *Response) UnmarshalJSON(b []byte) error {
	var inner struct {
		Success bool
		Message string
		Data    json.RawMessage
	}

	if err := json.Unmarshal(b, &inner); err != nil {
		return err
	}

	r.Success = inner.Success
	r.Message = inner.Message
	r.Data = inner.Data
	return nil
}

// GetPayload deposits the underlying Data payload into dest.
//
// This is an error if the structure was not unmarshalled from JSON using "encoding/json".
// You must check the Success flag before calling this method. It is an error to call this if Success is false.
func (r *Response) GetPayload(dest interface{}) error {
	raw, ok := r.Data.(json.RawMessage)
	if !ok {
		return errors.New("you should not use GetPayload() unless you have unmarshalled this structure from JSON")
	}

	if !r.Success {
		return errors.New("cannot use GetPayload() on a response that was not successful")
	}

	return json.Unmarshal(raw, dest)
}

// GetError is similar to GetPayload but for error responses.
//
// The same general constraints apply: you must check the Success flag, and this may only be used if you have unmarshalled the record.
func (r *Response) GetError(dest *error) error {
	raw, ok := r.Data.(json.RawMessage)
	if !ok {
		return errors.New("you should not use GetError() unless you have unmarshalled this structure from JSON")
	}

	if r.Success {
		return errors.New("cannot use GetError() on a response that was successful")
	}

	var ed ErrorData
	if err := json.Unmarshal(raw, &ed); err != nil {
		return err
	}

	*dest = fmt.Errorf("%s (code: %s)", ed.Message, ed.Code)
	return nil
}

// DataResponse returns a response that wraps the data in the correct format.
// Error is always nil to make returning from a Lambda less cumbersome.
func DataResponse(data interface{}) (Response, error) {
	return Response{Success: true, Data: data}, nil
}

var (
	// ErrCodeInvalidProvider indicates that the user supplied an unsupported provider.
	// The user MUST change their provider. The server will not accept the request without modification.
	ErrCodeInvalidProvider ErrorCode = "unsupported_provider"
	// ErrCodeUnspecified indicates that the reason for the operation failure was unknown.
	// The user MAY attempt resubmitting their request as-is, but there is no guarantee it will succeed.
	ErrCodeUnspecified ErrorCode = "unspecified"
	// ErrCodeUnableToDecrypt indicates the server was unable to decrypt the credentials the client provided.
	ErrCodeUnableToDecrypt ErrorCode = "decryption_failure"
	// ErrCodeInvalidCredentials indicates that the users credentials were incorrect
	ErrCodeInvalidCredentials ErrorCode = "invalid_credentials"
	// ErrCodeInternalServerError indicates that a server occurred within the server and the server could not continue.
	// The user cannot fix this issue. They MAY retry again.
	ErrCodeInternalServerError ErrorCode = "internal_server_error"
	// ErrCodeUnableToEncrypt indicates that the server was unable to encrypt the users credentials.
	ErrCodeUnableToEncrypt ErrorCode = "encryption_failure"
	// ErrBadRequest indicates that the user supplied data that was invalid
	ErrBadRequest ErrorCode = "bad_request"
)

// ErrorCode contains all of the recognised error codes in the KeyConjurer API.
type ErrorCode string

// ErrorData encapsulates error information relating to an AWS Lambda call.
// Lambda does not make it trivial to return HTTP status codes, so instead the application should interrogate the Code value in this struct.
type ErrorData struct {
	Code    ErrorCode
	Message string
}

// ErrorResponse creates a standardized error response with an error message from the server.
// It also always returns a nil error, simply to make returning from a Lambda less cumbersome.
func ErrorResponse(code ErrorCode, message string) (Response, error) {
	return Response{Success: false, Message: message, Data: ErrorData{Code: code, Message: message}}, nil
}
