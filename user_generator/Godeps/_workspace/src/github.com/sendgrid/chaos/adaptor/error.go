package adaptor

import (
	"errors"
	"net/http"
)

// Error is passed back to callers
type AdaptorError struct {
	Err                 error
	SuggestedStatusCode int
	Field               string
}

// Error matches the interface for errors
func (e *AdaptorError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

// NewError creates a default apidError with http.StatusInternalServerError
func NewError(msg string) *AdaptorError {
	return &AdaptorError{
		Err:                 errors.New(msg),
		SuggestedStatusCode: http.StatusInternalServerError,
	}
}

// NewErrorWithStatus lets you specify the status code
func NewErrorWithStatus(msg string, status int) *AdaptorError {
	return &AdaptorError{
		Err:                 errors.New(msg),
		SuggestedStatusCode: status,
	}
}
