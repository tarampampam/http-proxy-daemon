package main

// Server error for using in HTTP responses.
type ServerError struct {
	error
	Code    int    `json:"code"`
	IsError bool   `json:"error"`
	Message string `json:"message"`
}

// Server error constructor.
func NewServerError(code int, message string) *ServerError {
	return &ServerError{
		Code:    code,
		IsError: true,
		Message: message,
	}
}

// Get error message.
func (e *ServerError) Error() string {
	return e.Message
}
