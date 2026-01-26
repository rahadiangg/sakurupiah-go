package sakurupiah

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// VerifyAndParseCallback verifies the callback signature and parses the callback data.
//
// This is the RECOMMENDED method for handling incoming payment callbacks from Sakurupiah.
// It performs signature verification to ensure the callback is authentic and from Sakurupiah,
// then parses the JSON payload into a CallbackRequest struct.
//
// # Security
//
// Always verify callback signatures before processing payment notifications.
// Never trust callbacks without signature verification as they could be forged.
//
// # Example
//
//	func handlePaymentCallback(w http.ResponseWriter, r *http.Request) {
//	    body, _ := io.ReadAll(r.Body)
//
//	    callback, err := client.VerifyAndParseCallback(r.Header, body)
//	    if err != nil {
//	        http.Error(w, "Invalid signature", http.StatusBadRequest)
//	        return
//	    }
//
//	    // Process payment status
//	    if callback.Status == sakurupiah.StatusSuccess {
//	        // Update order status in your database
//	        updateOrderStatus(callback.MerchantRef, "paid")
//	    }
//
//	    // Send success response
//	    w.Header().Set("Content-Type", "application/json")
//	    json.NewEncoder(w).Encode(map[string]bool{"success": true})
//	}
//
// # Callback Headers
//
//   - X-Callback-Signature: HMAC-SHA256 signature of the JSON payload
//   - X-Callback-Event: Event type (always "payment_status")
//
// # Parameters
//
//   - headers: HTTP headers from the callback request
//   - body: Raw JSON bytes from the request body
//
// # Return
//
// *CallbackRequest containing the parsed callback data
func (c *Client) VerifyAndParseCallback(headers http.Header, body []byte) (*CallbackRequest, error) {
	// Extract headers
	callbackSig := headers.Get("X-Callback-Signature")
	callbackEvent := headers.Get("X-Callback-Event")

	if callbackSig == "" {
		return nil, fmt.Errorf("missing X-Callback-Signature header")
	}

	if callbackEvent != "payment_status" {
		return nil, fmt.Errorf("invalid callback event: %s", callbackEvent)
	}

	// Verify signature
	if !c.VerifyCallbackSignature(body, callbackSig) {
		return nil, ErrInvalidSignature
	}

	// Parse JSON
	var callback CallbackRequest
	if err := json.Unmarshal(body, &callback); err != nil {
		return nil, fmt.Errorf("failed to parse callback JSON: %w", err)
	}

	callback.RawPayload = body

	return &callback, nil
}

// HandleCallbackWithFunc processes a callback using a handler function.
//
// This is a convenience method that combines signature verification, parsing,
// and custom handler execution. It automatically generates an appropriate
// response message based on the payment status.
//
// # Example
//
//	func handlePaymentCallback(w http.ResponseWriter, r *http.Request) {
//	    body, _ := io.ReadAll(r.Body)
//
//	    resp, err := client.HandleCallbackWithFunc(r.Header, body, func(callback *sakurupiah.CallbackRequest) error {
//	        // Update your database
//	        return updateOrder(callback.MerchantRef, callback.Status)
//	    })
//
//	    if err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//
//	    json.NewEncoder(w).Encode(resp)
//	}
//
// # Parameters
//
//   - headers: HTTP headers from the callback request
//   - body: Raw JSON bytes from the request body
//   - handler: Function to process the verified callback
//
// # Return
//
// *CallbackResponse containing the success status and message
func (c *Client) HandleCallbackWithFunc(headers http.Header, body []byte, handler func(*CallbackRequest) error) (*CallbackResponse, error) {
	callback, err := c.VerifyAndParseCallback(headers, body)
	if err != nil {
		return &CallbackResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// Call the handler function
	if err := handler(callback); err != nil {
		return &CallbackResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// Determine message based on status
	var message string
	switch callback.Status {
	case StatusSuccess:
		message = "Payment status berhasil"
	case StatusExpired:
		message = "Payment status expired"
	case StatusPending:
		message = "Payment status pending"
	default:
		message = "Payment status processed"
	}

	return &CallbackResponse{
		Success: true,
		Message: message,
	}, nil
}

// SendCallbackResponse sends a JSON response to the callback.
//
// This is a helper function for sending properly formatted JSON responses
// to Sakurupiah callback requests.
//
// # Example
//
//	resp := sakurupiah.CallbackResponse{
//	    Success: true,
//	    Message: "Payment processed",
//	}
//	sakurupiah.SendCallbackResponse(w, resp)
//
// # Parameters
//
//   - w: http.ResponseWriter to write the response to
//   - resp: CallbackResponse containing the response data
func SendCallbackResponse(w http.ResponseWriter, resp CallbackResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CallbackHandlerFunc is a function type for handling callbacks.
//
// Custom handler functions should implement this signature to process
// verified callback requests.
type CallbackHandlerFunc func(*CallbackRequest) error

// NewCallbackHandler creates an HTTP handler function for processing callbacks.
//
// This is a convenient method that returns an http.HandlerFunc which handles
// signature verification, parsing, and custom processing. The returned handler
// can be used directly with http.HandleFunc or as a route handler.
//
// # Example
//
//	http.HandleFunc("/callback", client.NewCallbackHandler(func(callback *sakurupiah.CallbackRequest) error {
//	    // Process the payment callback
//	    if callback.Status == sakurupiah.StatusSuccess {
//	        // Update your database
//	        return updateOrderStatus(callback.MerchantRef, "paid")
//	    }
//	    return nil
//	}))
//
// # Features
//
//   - Automatic signature verification
//   - JSON parsing and error handling
//   - Method validation (POST only)
//   - Automatic response generation
//
// # Parameters
//
//   - handler: Function to process verified callbacks
//
// # Return
//
// http.HandlerFunc that can be used as an HTTP handler
func (c *Client) NewCallbackHandler(handler CallbackHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			SendCallbackResponse(w, CallbackResponse{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		// Read request body
		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			SendCallbackResponse(w, CallbackResponse{
				Success: false,
				Message: "Failed to read request body",
			})
			return
		}

		// Process callback
		resp, err := c.HandleCallbackWithFunc(r.Header, body, handler)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			SendCallbackResponse(w, *resp)
			return
		}

		SendCallbackResponse(w, *resp)
	}
}

// readRequestBody reads the request body.
//
// Helper function to safely read and parse the HTTP request body.
func readRequestBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	return body, err
}

// ============================================================
// Callback Handler Builder
// ============================================================

// CallbackHandlerBuilder helps build a callback handler with validation.
//
// The builder pattern provides a fluent API for creating HTTP handlers
// with separate callbacks for different payment statuses (success, expired, pending).
// This is useful for handling different payment scenarios with dedicated logic.
//
// # Example
//
//	handler := sakurupiah.NewCallbackHandlerBuilder(client).
//	    OnSuccess(func(cb *sakurupiah.CallbackRequest) error {
//	        // Fulfill order
//	        return fulfillOrder(cb.MerchantRef)
//	    }).
//	    OnExpired(func(cb *sakurupiah.CallbackRequest) error {
//	        // Handle expired payment
//	        return cancelOrder(cb.MerchantRef)
//	    }).
//	    Build()
//
//	http.HandleFunc("/callback", handler)
type CallbackHandlerBuilder struct {
	client    *Client
	onSuccess func(*CallbackRequest) error
	onExpired func(*CallbackRequest) error
	onPending func(*CallbackRequest) error
	onError   func(*CallbackRequest, error) error
}

// NewCallbackHandlerBuilder creates a new callback handler builder.
//
// Use this to create a builder for constructing HTTP handlers with separate
// callbacks for different payment statuses.
//
// # Parameters
//
//   - client: The Sakurupiah client for signature verification
//
// # Return
//
// *CallbackHandlerBuilder for building the handler
func NewCallbackHandlerBuilder(client *Client) *CallbackHandlerBuilder {
	return &CallbackHandlerBuilder{
		client: client,
	}
}

// OnSuccess sets the handler for successful payments.
//
// The provided handler will be called when a payment with status "berhasil" (success)
// is received in the callback.
//
// # Parameters
//
//   - handler: Function to handle successful payment callbacks
//
// # Return
//
// *CallbackHandlerBuilder for method chaining
func (b *CallbackHandlerBuilder) OnSuccess(handler func(*CallbackRequest) error) *CallbackHandlerBuilder {
	b.onSuccess = handler
	return b
}

// OnExpired sets the handler for expired payments.
//
// The provided handler will be called when a payment with status "expired"
// is received in the callback.
//
// # Parameters
//
//   - handler: Function to handle expired payment callbacks
//
// # Return
//
// *CallbackHandlerBuilder for method chaining
func (b *CallbackHandlerBuilder) OnExpired(handler func(*CallbackRequest) error) *CallbackHandlerBuilder {
	b.onExpired = handler
	return b
}

// OnPending sets the handler for pending payments.
//
// The provided handler will be called when a payment with status "pending"
// is received in the callback.
//
// # Parameters
//
//   - handler: Function to handle pending payment callbacks
//
// # Return
//
// *CallbackHandlerBuilder for method chaining
func (b *CallbackHandlerBuilder) OnPending(handler func(*CallbackRequest) error) *CallbackHandlerBuilder {
	b.onPending = handler
	return b
}

// OnError sets the error handler.
//
// The provided handler will be called when an error occurs during signature
// verification, parsing, or when a status handler returns an error.
//
// # Parameters
//
//   - handler: Function to handle errors (receives callback and error)
//
// # Return
//
// *CallbackHandlerBuilder for method chaining
func (b *CallbackHandlerBuilder) OnError(handler func(*CallbackRequest, error) error) *CallbackHandlerBuilder {
	b.onError = handler
	return b
}

// Build creates the HTTP handler function.
//
// Builds the final http.HandlerFunc from the configured builder.
// The handler will route callbacks to the appropriate status handler
// based on the payment status in the callback.
//
// # Return
//
// http.HandlerFunc ready to use with http.HandleFunc or as a route handler
func (b *CallbackHandlerBuilder) Build() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			SendCallbackResponse(w, CallbackResponse{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		body, err := readRequestBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			SendCallbackResponse(w, CallbackResponse{
				Success: false,
				Message: "Failed to read request body",
			})
			return
		}

		callback, err := b.client.VerifyAndParseCallback(r.Header, body)
		if err != nil {
			if b.onError != nil {
				b.onError(nil, err)
			}
			w.WriteHeader(http.StatusBadRequest)
			SendCallbackResponse(w, CallbackResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		// Route to appropriate handler based on status
		var handlerErr error
		switch callback.Status {
		case StatusSuccess:
			if b.onSuccess != nil {
				handlerErr = b.onSuccess(callback)
			}
		case StatusExpired:
			if b.onExpired != nil {
				handlerErr = b.onExpired(callback)
			}
		case StatusPending:
			if b.onPending != nil {
				handlerErr = b.onPending(callback)
			}
		}

		if handlerErr != nil && b.onError != nil {
			b.onError(callback, handlerErr)
		}

		// Send response
		var resp CallbackResponse
		if handlerErr != nil {
			resp = CallbackResponse{
				Success: false,
				Message: handlerErr.Error(),
			}
			w.WriteHeader(http.StatusBadRequest)
		} else {
			var message string
			switch callback.Status {
			case StatusSuccess:
				message = "Payment status berhasil"
			case StatusExpired:
				message = "Payment status expired"
			case StatusPending:
				message = "Payment status pending"
			default:
				message = "Payment status processed"
			}
			resp = CallbackResponse{
				Success: true,
				Message: message,
			}
		}

		SendCallbackResponse(w, resp)
	}
}
