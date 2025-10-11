package common

// ErrorResponse represents a standard error response body.
// swagger:model ErrorResponse
type ErrorResponse struct {
    // example: user not found
    Error string `json:"error"`
}