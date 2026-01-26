// Package sakurupiah provides a Go SDK for integrating with the Sakurupiah Payment Gateway.
//
// The SDK supports creating payment invoices, listing payment channels, checking balance,
// querying transaction history, and handling payment callbacks with signature verification.
//
// # Basic Usage
//
//	import sakurupiah "github.com/rahadiangg/sakururupiah-go/sakurupiah"
//
//	client, err := sakurupiah.NewClient(sakurupiah.Config{
//	    APIID:    "YOUR_API_ID",
//	    APIKey:   "YOUR_API_KEY",
//	    IsSandbox: true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create an invoice
//	resp, err := client.CreateInvoiceSimple(
//	    "QRIS", "John Doe", "628123456789",
//	    10000, "INV-001", callbackURL, returnURL,
//	)
//
// # Environment Variables
//
// Use environment variables to securely store credentials:
//
//	export SAKURUPIAH_API_ID="your-api-id"
//	export SAKURUPIAH_API_KEY="your-api-key"
//
// # Callback Handling
//
// The SDK provides secure callback handling with automatic signature verification.
// Always verify callbacks using VerifyAndParseCallback() before processing payments.
package sakurupiah

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// ProductionBaseURL is the base URL for production environment
	ProductionBaseURL = "https://sakurupiah.id/api/"
	// SandboxBaseURL is the base URL for sandbox environment
	SandboxBaseURL = "https://sakurupiah.id/api-sanbox/"
	// DefaultTimeout is the default HTTP timeout
	DefaultTimeout = 30 * time.Second
)

// Client represents the Sakurupiah API client.
// It handles authentication, signature generation, and HTTP communication with the Sakurupiah API.
//
// # Environment
//
// The client can operate in two environments:
//   - Production: Uses https://sakurupiah.id/api/
//   - Sandbox: Uses https://sakurupiah.id/api-sanbox/
//
// # Authentication
//
// The client uses Bearer token authentication with your API Key.
// Set IsSandbox to true when developing and testing.
type Client struct {
	apiID       string       // API ID obtained from Sakurupiah dashboard
	apiKey      string       // API Key obtained from Sakurupiah dashboard
	baseURL     string       // Base URL for API requests (production or sandbox)
	httpClient  *http.Client // HTTP client for making requests
	callbackURL string       // Default callback URL for payment notifications
	returnURL   string       // Default return URL for redirect after payment
}

// Config holds the configuration for creating a new Client.
//
// # Required Fields
//   - APIID: Your Sakurupiah API ID from the dashboard
//   - APIKey: Your Sakurupiah API Key from the dashboard
//
// # Optional Fields
//   - IsSandbox: Set to true to use sandbox environment (default: false)
//   - Timeout: HTTP request timeout (default: 30 seconds)
//   - HTTPClient: Custom HTTP client (default: nil creates a new client)
//   - DefaultCallbackURL: Default callback URL for all invoices
//   - DefaultReturnURL: Default return URL for all invoices
type Config struct {
	// API ID is obtained from Sakurupiah dashboard
	APIID string
	// API Key is obtained from Sakurupiah dashboard
	APIKey string
	// IsSandbox indicates whether to use sandbox environment
	// Set to true for development and testing, false for production
	IsSandbox bool
	// Timeout is the HTTP request timeout (default: 30s)
	Timeout time.Duration
	// HTTPClient is a custom HTTP client (optional)
	// If nil, a new client with the configured timeout will be created
	HTTPClient *http.Client
	// DefaultCallbackURL is the default callback URL for transactions
	// If set, invoices created without a callback URL will use this value
	DefaultCallbackURL string
	// DefaultReturnURL is the default return URL for transactions
	// If set, invoices created without a return URL will use this value
	DefaultReturnURL string
}

// NewClient creates a new Sakurupiah API client with the given configuration.
//
// The client validates that APIID and APIKey are provided and sets up the appropriate
// base URL based on the IsSandbox flag.
//
// # Example
//
//	client, err := sakurupiah.NewClient(sakurupiah.Config{
//	    APIID:    "SANBOX-12345",
//	    APIKey:   "SANBOX-abcde12345",
//	    IsSandbox: true,
//	    Timeout:  60 * time.Second,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Environment
//
// Use environment variables for better security:
//
//	client, err := sakurupiah.NewClient(sakurupiah.Config{
//	    APIID:    os.Getenv("SAKURUPIAH_API_ID"),
//	    APIKey:   os.Getenv("SAKURUPIAH_API_KEY"),
//	    IsSandbox: os.Getenv("ENV") != "production",
//	})
func NewClient(cfg Config) (*Client, error) {
	if cfg.APIID == "" {
		return nil, ErrMissingAPIID
	}
	if cfg.APIKey == "" {
		return nil, ErrMissingAPIKey
	}

	baseURL := ProductionBaseURL
	if cfg.IsSandbox {
		baseURL = SandboxBaseURL
	}

	timeout := DefaultTimeout
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	return &Client{
		apiID:       cfg.APIID,
		apiKey:      cfg.APIKey,
		baseURL:     baseURL,
		httpClient:  httpClient,
		callbackURL: cfg.DefaultCallbackURL,
		returnURL:   cfg.DefaultReturnURL,
	}, nil
}

// GetAPIID returns the client's API ID.
// This is the API ID obtained from the Sakurupiah dashboard.
func (c *Client) GetAPIID() string {
	return c.apiID
}

// GetAPIKey returns the client's API Key.
// This is the API Key obtained from the Sakurupiah dashboard.
// Use this carefully and never expose it in client-side code or logs.
func (c *Client) GetAPIKey() string {
	return c.apiKey
}

// IsSandbox returns true if using sandbox environment.
// Sandbox environment is used for development and testing.
// Production mode should be used for live transactions.
func (c *Client) IsSandbox() bool {
	return strings.Contains(c.baseURL, "api-sanbox")
}

// SetDefaultCallbackURL sets the default callback URL for all invoices.
// If set, invoices created without a callback URL will use this value.
// This is useful for avoiding repetitive callback URL configuration.
func (c *Client) SetDefaultCallbackURL(callbackURL string) {
	c.callbackURL = callbackURL
}

// SetDefaultReturnURL sets the default return URL for all invoices.
// If set, invoices created without a return URL will use this value.
// This is useful for avoiding repetitive return URL configuration.
func (c *Client) SetDefaultReturnURL(returnURL string) {
	c.returnURL = returnURL
}

// GenerateSignature creates an HMAC-SHA256 signature for invoice creation.
//
// The signature is generated from: apiID + method + merchantRef + amount
//
// This signature is required for all invoice creation requests to ensure
// the integrity of the transaction data and authenticate the merchant.
//
// # Algorithm
//
// signature = HMAC-SHA256(apiID + method + merchantRef + amount, apiKey)
//
// # Example
//
//	sig := client.GenerateSignature("QRIS", "INV-001", 10000)
//	// Returns: "a1b2c3d4e5f6..."
//
// # Parameters
//
//   - method: Payment channel code (e.g., "QRIS", "BCAVA", "DANA")
//   - merchantRef: Your unique reference/transaction ID
//   - amount: Transaction amount in IDR
//
// # Return
//
// 64-character hexadecimal string representing the HMAC-SHA256 signature
func (c *Client) GenerateSignature(method, merchantRef string, amount int64) string {
	data := fmt.Sprintf("%s%s%s%d", c.apiID, method, merchantRef, amount)
	h := hmac.New(sha256.New, []byte(c.apiKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyCallbackSignature verifies the signature from a payment callback.
//
// The callback signature is generated from: JSON payload + API key
//
// This method verifies that the callback payload was not tampered with
// and that it came from the Sakurupiah servers.
//
// # Algorithm
//
// signature = HMAC-SHA256(rawJSONPayload, apiKey)
//
// # Example
//
//	body, _ := io.ReadAll(r.Body)
//	valid := client.VerifyCallbackSignature(body, signatureFromHeader)
//	if !valid {
//	    // Reject the callback
//	    http.Error(w, "Invalid signature", http.StatusBadRequest)
//	    return
//	}
//
// # Security
//
// Always verify callback signatures before processing payment notifications.
// Never trust callbacks without signature verification as they could be forged.
//
// # Parameters
//
//   - jsonPayload: Raw JSON bytes received in the callback request body
//   - receivedSignature: Signature from X-Callback-Signature header
//
// # Return
//
// true if signature is valid, false otherwise
func (c *Client) VerifyCallbackSignature(jsonPayload []byte, receivedSignature string) bool {
	h := hmac.New(sha256.New, []byte(c.apiKey))
	h.Write(jsonPayload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(receivedSignature))
}

// doRequest performs an HTTP request with Bearer token authentication
func (c *Client) doRequest(endpoint string, formData url.Values) (*http.Response, error) {
	reqURL := c.baseURL + endpoint

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.httpClient.Do(req)
}

// doJSONRequest performs an HTTP request and returns the JSON response
func (c *Client) doJSONRequest(endpoint string, formData url.Values, result interface{}) error {
	resp, err := c.doRequest(endpoint, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp.StatusCode, body)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

// handleErrorResponse processes error responses from the API
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return &APIError{
			StatusCode: statusCode,
			Message:    string(body),
		}
	}

	return &APIError{
		StatusCode: statusCode,
		Status:     errResp.Status,
		Message:    errResp.Message,
	}
}

// buildFormData converts a map to url.Values
func buildFormData(data map[string]string) url.Values {
	values := url.Values{}
	for key, value := range data {
		values.Set(key, value)
	}
	return values
}

// buildFormDataWithArrays converts a map with array values to url.Values
func buildFormDataWithArrays(data map[string]string, arrays map[string][]string) url.Values {
	values := url.Values{}
	for key, value := range data {
		values.Set(key, value)
	}
	for key, slice := range arrays {
		for _, value := range slice {
			values.Add(key, value)
		}
	}
	return values
}

// doJSONRequestWithArray performs an HTTP request with array fields
func (c *Client) doJSONRequestWithArray(endpoint string, data map[string]string, arrays map[string][]string, result interface{}) error {
	formData := buildFormDataWithArrays(data, arrays)

	reqURL := c.baseURL + endpoint
	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp.StatusCode, body)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

// postRawJSON performs a POST request with raw JSON body
// Note: Currently unused but kept for potential future use
//
//nolint:unused // Kept for potential future API endpoints that may require JSON body
func (c *Client) postRawJSON(endpoint string, jsonBody []byte) (*http.Response, error) {
	reqURL := c.baseURL + endpoint
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}
