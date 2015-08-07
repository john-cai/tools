package client

const (
	ErrorMessageInternalService  = "An internal service had an error. This incident has been logged"
	ErrorMessageBadPayload       = "Bad json payload"
	ErrorMessageResourceNotFound = "Resource not found"
	ErrorMessageForbidden        = "You don't have permission to access this resource"
)

// ChaosErrorResult is the struct clients should consume
type ChaosErrorResult struct {
	Errors []ChaosError `json:"errors"`
}

// ChaosError represents individual errors the client may get
type ChaosError struct {
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}
