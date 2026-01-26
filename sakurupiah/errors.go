package sakurupiah

import "fmt"

// Error definitions
var (
	// ErrMissingAPIID is returned when API ID is not provided
	ErrMissingAPIID = fmt.Errorf("API ID is required")
	// ErrMissingAPIKey is returned when API Key is not provided
	ErrMissingAPIKey = fmt.Errorf("API Key is required")
	// ErrInvalidAmount is returned when amount is invalid
	ErrInvalidAmount = fmt.Errorf("invalid amount")
	// ErrInvalidPhone is returned when phone number is invalid
	ErrInvalidPhone = fmt.Errorf("invalid phone number")
	// ErrMissingMerchantRef is returned when merchant reference is missing
	ErrMissingMerchantRef = fmt.Errorf("merchant reference is required")
	// ErrMissingMethod is returned when payment method is missing
	ErrMissingMethod = fmt.Errorf("payment method is required")
	// ErrInvalidSignature is returned when signature verification fails
	ErrInvalidSignature = fmt.Errorf("invalid signature")
)

// APIError represents an error response from the API
type APIError struct {
	StatusCode int
	Status     string
	Message    string
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Status != "" {
		return fmt.Sprintf("API error (status %s): %s", e.Status, e.Message)
	}
	return fmt.Sprintf("API error (code %d): %s", e.StatusCode, e.Message)
}

// ErrorResponse represents the standard error response structure
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
