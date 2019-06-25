package keyconjurer

// Response is the generic structure of the lambda responses
type Response struct {
	Success bool
	Message string
	Data    interface{}
}

// CreateResponseUnexpectedError is used when unexpected errors occur
func CreateResponseUnexpectedError() *Response {
	return CreateResponseError("Unexpected error occurred")
}

// CreateResponseError creates a general usage error response
func CreateResponseError(reason string) *Response {
	return &Response{
		Success: false,
		Message: reason,
		Data:    nil}

}

// CreateResponseSuccess creates a general usage success response
func CreateResponseSuccess(data interface{}) *Response {
	return &Response{
		Success: true,
		Message: "success",
		Data:    data}
}
