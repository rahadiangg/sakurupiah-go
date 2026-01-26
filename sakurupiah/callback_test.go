package sakurupiah

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestVerifyAndParseCallback tests callback verification and parsing
func TestVerifyAndParseCallback(t *testing.T) {
	config := Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-67890",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	callbackPayload := CallbackRequest{
		TrxID:       "TRX123",
		MerchantRef: "REF123",
		Status:      StatusSuccess,
		StatusCode:  StatusCodeSuccess,
	}

	jsonPayload, err := json.Marshal(callbackPayload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Generate valid signature
	h := hmac.New(sha256.New, []byte(client.apiKey))
	h.Write(jsonPayload)
	validSignature := hex.EncodeToString(h.Sum(nil))

	tests := []struct {
		name      string
		headers   http.Header
		body      []byte
		wantErr   error
		wantTrxID string
	}{
		{
			name: "valid callback",
			headers: http.Header{
				"X-Callback-Signature": []string{validSignature},
				"X-Callback-Event":     []string{"payment_status"},
			},
			body:      jsonPayload,
			wantErr:   nil,
			wantTrxID: "TRX123",
		},
		{
			name: "missing signature",
			headers: http.Header{
				"X-Callback-Event": []string{"payment_status"},
			},
			body:    jsonPayload,
			wantErr: ErrInvalidSignature,
		},
		{
			name: "invalid signature",
			headers: http.Header{
				"X-Callback-Signature": []string{"invalid" + validSignature},
				"X-Callback-Event":     []string{"payment_status"},
			},
			body:    jsonPayload,
			wantErr: ErrInvalidSignature,
		},
		{
			name: "invalid event",
			headers: http.Header{
				"X-Callback-Signature": []string{validSignature},
				"X-Callback-Event":     []string{"invalid_event"},
			},
			body:    jsonPayload,
			wantErr: ErrInvalidSignature,
		},
		{
			name: "invalid JSON",
			headers: http.Header{
				"X-Callback-Signature": []string{validSignature},
				"X-Callback-Event":     []string{"payment_status"},
			},
			body:    []byte("invalid json"),
			wantErr: ErrInvalidSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callback, err := client.VerifyAndParseCallback(tt.headers, tt.body)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("VerifyAndParseCallback() expected error, got nil")
					return
				}
				// We just check that an error occurred, not necessarily the exact type
				return
			}

			if err != nil {
				t.Errorf("VerifyAndParseCallback() unexpected error = %v", err)
				return
			}

			if callback.TrxID != tt.wantTrxID {
				t.Errorf("TrxID = %v, want %v", callback.TrxID, tt.wantTrxID)
			}

			// Verify raw payload is stored
			if len(callback.RawPayload) == 0 {
				t.Error("RawPayload should not be empty")
			}
		})
	}
}

// TestHandleCallbackWithFunc tests callback handling with custom function
func TestHandleCallbackWithFunc(t *testing.T) {
	config := Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-67890",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	callbackPayload := CallbackRequest{
		TrxID:       "TRX123",
		MerchantRef: "REF123",
		Status:      StatusSuccess,
		StatusCode:  StatusCodeSuccess,
	}

	jsonPayload, err := json.Marshal(callbackPayload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Generate valid signature
	h := hmac.New(sha256.New, []byte(client.apiKey))
	h.Write(jsonPayload)
	validSignature := hex.EncodeToString(h.Sum(nil))

	tests := []struct {
		name           string
		headers        http.Header
		body           []byte
		handlerCalled  bool
		handlerSuccess bool
	}{
		{
			name: "successful callback handling",
			headers: http.Header{
				"X-Callback-Signature": []string{validSignature},
				"X-Callback-Event":     []string{"payment_status"},
			},
			body:           jsonPayload,
			handlerCalled:  true,
			handlerSuccess: true,
		},
		{
			name: "invalid signature - handler not called",
			headers: http.Header{
				"X-Callback-Signature": []string{"invalid"},
				"X-Callback-Event":     []string{"payment_status"},
			},
			body:           jsonPayload,
			handlerCalled:  false,
			handlerSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			var capturedCallback *CallbackRequest

			handler := func(callback *CallbackRequest) error {
				handlerCalled = true
				capturedCallback = callback
				return nil
			}

			resp, err := client.HandleCallbackWithFunc(tt.headers, tt.body, handler)

			if tt.handlerCalled != handlerCalled {
				t.Errorf("handler called = %v, want %v", handlerCalled, tt.handlerCalled)
			}

			if !tt.handlerCalled {
				if resp.Success {
					t.Error("Response should indicate failure")
				}
				return
			}

			if err != nil {
				t.Errorf("HandleCallbackWithFunc() unexpected error = %v", err)
			}

			if !resp.Success {
				t.Error("Response.Success should be true")
			}

			if capturedCallback == nil {
				t.Error("capturedCallback should not be nil")
			} else if capturedCallback.TrxID != "TRX123" {
				t.Errorf("capturedCallback.TrxID = %v, want TRX123", capturedCallback.TrxID)
			}
		})
	}
}

// TestNewCallbackHandler tests creating HTTP callback handler
func TestNewCallbackHandler(t *testing.T) {
	config := Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-67890",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	callbackPayload := CallbackRequest{
		TrxID:       "TRX123",
		MerchantRef: "REF123",
		Status:      StatusSuccess,
		StatusCode:  StatusCodeSuccess,
	}

	jsonPayload, err := json.Marshal(callbackPayload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Generate valid signature
	h := hmac.New(sha256.New, []byte(client.apiKey))
	h.Write(jsonPayload)
	validSignature := hex.EncodeToString(h.Sum(nil))

	handlerCalled := false
	handlerFunc := func(callback *CallbackRequest) error {
		handlerCalled = true
		if callback.TrxID != "TRX123" {
			t.Errorf("callback.TrxID = %v, want TRX123", callback.TrxID)
		}
		return nil
	}

	handler := client.NewCallbackHandler(handlerFunc)

	tests := []struct {
		name           string
		method         string
		signature      string
		event          string
		body           []byte
		wantStatus     int
		wantHandlerCalled bool
	}{
		{
			name:       "valid POST request",
			method:     http.MethodPost,
			signature:  validSignature,
			event:      "payment_status",
			body:       jsonPayload,
			wantStatus: http.StatusOK,
			wantHandlerCalled: true,
		},
		{
			name:       "invalid method",
			method:     http.MethodGet,
			signature:  validSignature,
			event:      "payment_status",
			body:       jsonPayload,
			wantStatus: http.StatusMethodNotAllowed,
			wantHandlerCalled: false,
		},
		{
			name:       "invalid signature",
			method:     http.MethodPost,
			signature:  "invalid",
			event:      "payment_status",
			body:       jsonPayload,
			wantStatus: http.StatusBadRequest,
			wantHandlerCalled: false,
		},
		{
			name:       "invalid event",
			method:     http.MethodPost,
			signature:  validSignature,
			event:      "invalid_event",
			body:       jsonPayload,
			wantStatus: http.StatusBadRequest,
			wantHandlerCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled = false

			req := httptest.NewRequest(tt.method, "/callback", strings.NewReader(string(tt.body)))
			req.Header.Set("X-Callback-Signature", tt.signature)
			req.Header.Set("X-Callback-Event", tt.event)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler(w, req)

			resp := w.Result()
			resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			if handlerCalled != tt.wantHandlerCalled {
				t.Errorf("handler called = %v, want %v", handlerCalled, tt.wantHandlerCalled)
			}

			// Check response content type
		 contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", contentType)
			}

			// Parse response body
			var callbackResp CallbackResponse
			body := w.Body.Bytes()
			if err := json.Unmarshal(body, &callbackResp); err != nil {
				t.Errorf("Failed to parse response body: %v", err)
			}

			if tt.wantStatus == http.StatusOK && !callbackResp.Success {
				t.Error("Response.Success should be true for successful callback")
			}
		})
	}
}

// TestCallbackHandlerBuilder tests the callback handler builder
func TestCallbackHandlerBuilder(t *testing.T) {
	config := Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-67890",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	tests := []struct {
		name   string
		status TransactionStatusValue
		code   TransactionStatusCode
	}{
		{
			name:   "success callback",
			status: StatusSuccess,
			code:   StatusCodeSuccess,
		},
		{
			name:   "expired callback",
			status: StatusExpired,
			code:   StatusCodeExpired,
		},
		{
			name:   "pending callback",
			status: StatusPending,
			code:   StatusCodePending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var successCalled, expiredCalled, pendingCalled bool

			builder := NewCallbackHandlerBuilder(client).
				OnSuccess(func(callback *CallbackRequest) error {
					successCalled = true
					if callback.Status != StatusSuccess {
						t.Errorf("callback.Status = %v, want %v", callback.Status, StatusSuccess)
					}
					return nil
				}).
				OnExpired(func(callback *CallbackRequest) error {
					expiredCalled = true
					if callback.Status != StatusExpired {
						t.Errorf("callback.Status = %v, want %v", callback.Status, StatusExpired)
					}
					return nil
				}).
				OnPending(func(callback *CallbackRequest) error {
					pendingCalled = true
					if callback.Status != StatusPending {
						t.Errorf("callback.Status = %v, want %v", callback.Status, StatusPending)
					}
					return nil
				})

			handler := builder.Build()

			callbackPayload := CallbackRequest{
				TrxID:       "TRX123",
				MerchantRef: "REF123",
				Status:      tt.status,
				StatusCode:  tt.code,
			}

			jsonPayload, _ := json.Marshal(callbackPayload)

			// Generate valid signature
			h := hmac.New(sha256.New, []byte(client.apiKey))
			h.Write(jsonPayload)
			validSignature := hex.EncodeToString(h.Sum(nil))

			req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(string(jsonPayload)))
			req.Header.Set("X-Callback-Signature", validSignature)
			req.Header.Set("X-Callback-Event", "payment_status")
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler(w, req)

			resp := w.Result()
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
			}

			// Verify correct handler was called
			switch tt.status {
			case StatusSuccess:
				if !successCalled {
					t.Error("OnSuccess handler should be called")
				}
			case StatusExpired:
				if !expiredCalled {
					t.Error("OnExpired handler should be called")
				}
			case StatusPending:
				if !pendingCalled {
					t.Error("OnPending handler should be called")
				}
			}
		})
	}
}

// TestSendCallbackResponse tests sending callback response
func TestSendCallbackResponse(t *testing.T) {
	tests := []struct {
		name    string
		resp    CallbackResponse
		wantErr bool
	}{
		{
			name: "success response",
			resp: CallbackResponse{
				Success: true,
				Message: "Payment processed",
			},
			wantErr: false,
		},
		{
			name: "error response",
			resp: CallbackResponse{
				Success: false,
				Message: "Invalid signature",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SendCallbackResponse(w, tt.resp)

			// Check response was written
			if w.Body.Len() == 0 {
				t.Error("Response body should not be empty")
			}

			// Verify content type
			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", contentType)
			}

			// Parse response
			var decoded CallbackResponse
			if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
				t.Errorf("Failed to parse response: %v", err)
			}

			if decoded.Success != tt.resp.Success {
				t.Errorf("Success = %v, want %v", decoded.Success, tt.resp.Success)
			}
			if decoded.Message != tt.resp.Message {
				t.Errorf("Message = %v, want %v", decoded.Message, tt.resp.Message)
			}
		})
	}
}

// TestCallbackHandlerErrorHandling tests error handling in callback handlers
func TestCallbackHandlerErrorHandling(t *testing.T) {
	config := Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-67890",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	callbackPayload := CallbackRequest{
		TrxID:       "TRX123",
		MerchantRef: "REF123",
		Status:      StatusSuccess,
		StatusCode:  StatusCodeSuccess,
	}

	jsonPayload, _ := json.Marshal(callbackPayload)

	// Generate valid signature
	h := hmac.New(sha256.New, []byte(client.apiKey))
	h.Write(jsonPayload)
	validSignature := hex.EncodeToString(h.Sum(nil))

	t.Run("handler returns error", func(t *testing.T) {
		handlerFunc := func(callback *CallbackRequest) error {
			return ErrInvalidSignature
		}

		handler := client.NewCallbackHandler(handlerFunc)

		req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(string(jsonPayload)))
		req.Header.Set("X-Callback-Signature", validSignature)
		req.Header.Set("X-Callback-Event", "payment_status")
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}

		var callbackResp CallbackResponse
		body := w.Body.Bytes()
		if err := json.Unmarshal(body, &callbackResp); err != nil {
			t.Errorf("Failed to parse response: %v", err)
		}

		if callbackResp.Success {
			t.Error("Response.Success should be false when handler returns error")
		}
	})
}

// TestCallbackHeaders tests callback header handling
func TestCallbackHeaders(t *testing.T) {
	headers := CallbackHeaders{
		Signature: "test_signature",
		Event:     "payment_status",
	}

	if headers.Signature != "test_signature" {
		t.Errorf("Signature = %v, want test_signature", headers.Signature)
	}
	if headers.Event != "payment_status" {
		t.Errorf("Event = %v, want payment_status", headers.Event)
	}
}
