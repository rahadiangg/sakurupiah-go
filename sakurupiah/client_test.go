package sakurupiah

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"
)

// TestNewClient tests creating a new client
func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "valid config - production",
			config: Config{
				APIID:     "TEST-12345",
				APIKey:    "test-key-12345",
				IsSandbox: false,
			},
			wantErr: nil,
		},
		{
			name: "valid config - sandbox",
			config: Config{
				APIID:     "SANBOX-12345",
				APIKey:    "sandbox-key-12345",
				IsSandbox: true,
			},
			wantErr: nil,
		},
		{
			name: "valid config with custom timeout",
			config: Config{
				APIID:     "TEST-12345",
				APIKey:    "test-key-12345",
				IsSandbox: false,
				Timeout:   60 * time.Second,
			},
			wantErr: nil,
		},
		{
			name: "valid config with default URLs",
			config: Config{
				APIID:              "TEST-12345",
				APIKey:             "test-key-12345",
				IsSandbox:          false,
				DefaultCallbackURL: "https://example.com/callback",
				DefaultReturnURL:   "https://example.com/return",
			},
			wantErr: nil,
		},
		{
			name: "missing API ID",
			config: Config{
				APIKey:    "test-key-12345",
				IsSandbox: false,
			},
			wantErr: ErrMissingAPIID,
		},
		{
			name: "missing API Key",
			config: Config{
				APIID:     "TEST-12345",
				IsSandbox: false,
			},
			wantErr: ErrMissingAPIKey,
		},
		{
			name:    "empty config",
			config:  Config{},
			wantErr: ErrMissingAPIID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				}
				if client != nil {
					t.Error("NewClient() should return nil client on error")
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error = %v", err)
				return
			}

			if client == nil {
				t.Fatal("NewClient() returned nil client")
			}

			// Verify client properties
			if client.apiID != tt.config.APIID {
				t.Errorf("client.apiID = %v, want %v", client.apiID, tt.config.APIID)
			}
			if client.apiKey != tt.config.APIKey {
				t.Errorf("client.apiKey = %v, want %v", client.apiKey, tt.config.APIKey)
			}

			// Verify base URL
			expectedURL := ProductionBaseURL
			if tt.config.IsSandbox {
				expectedURL = SandboxBaseURL
			}
			if client.baseURL != expectedURL {
				t.Errorf("client.baseURL = %v, want %v", client.baseURL, expectedURL)
			}

			// Verify default URLs
			if tt.config.DefaultCallbackURL != "" && client.callbackURL != tt.config.DefaultCallbackURL {
				t.Errorf("client.callbackURL = %v, want %v", client.callbackURL, tt.config.DefaultCallbackURL)
			}
			if tt.config.DefaultReturnURL != "" && client.returnURL != tt.config.DefaultReturnURL {
				t.Errorf("client.returnURL = %v, want %v", client.returnURL, tt.config.DefaultReturnURL)
			}
		})
	}
}

// TestClientGetters tests client getter methods
func TestClientGetters(t *testing.T) {
	config := Config{
		APIID:              "TEST-12345",
		APIKey:             "test-key-12345",
		IsSandbox:          true,
		DefaultCallbackURL: "https://example.com/callback",
		DefaultReturnURL:   "https://example.com/return",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.GetAPIID() != config.APIID {
		t.Errorf("GetAPIID() = %v, want %v", client.GetAPIID(), config.APIID)
	}

	if client.GetAPIKey() != config.APIKey {
		t.Errorf("GetAPIKey() = %v, want %v", client.GetAPIKey(), config.APIKey)
	}

	if !client.IsSandbox() {
		t.Error("IsSandbox() = false, want true")
	}
}

// TestClientSetters tests client setter methods
func TestClientSetters(t *testing.T) {
	config := Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	newCallbackURL := "https://new-example.com/callback"
	newReturnURL := "https://new-example.com/return"
	newPaymentMethod := MethodQRIS

	client.SetDefaultCallbackURL(newCallbackURL)
	if client.callbackURL != newCallbackURL {
		t.Errorf("SetDefaultCallbackURL() = %v, want %v", client.callbackURL, newCallbackURL)
	}

	client.SetDefaultReturnURL(newReturnURL)
	if client.returnURL != newReturnURL {
		t.Errorf("SetDefaultReturnURL() = %v, want %v", client.returnURL, newReturnURL)
	}

	client.SetDefaultPaymentMethod(newPaymentMethod)
	if client.defaultPaymentMethod != newPaymentMethod {
		t.Errorf("SetDefaultPaymentMethod() = %v, want %v", client.defaultPaymentMethod, newPaymentMethod)
	}

	if client.GetDefaultPaymentMethod() != newPaymentMethod {
		t.Errorf("GetDefaultPaymentMethod() = %v, want %v", client.GetDefaultPaymentMethod(), newPaymentMethod)
	}
}

// TestDefaultPaymentMethod tests default payment method configuration
func TestDefaultPaymentMethod(t *testing.T) {
	tests := []struct {
		name                  string
		config                Config
		expectedPaymentMethod string
	}{
		{
			name: "with default payment method",
			config: Config{
				APIID:                "TEST-12345",
				APIKey:               "test-key-12345",
				DefaultPaymentMethod: MethodQRIS,
			},
			expectedPaymentMethod: MethodQRIS,
		},
		{
			name: "with different default payment method",
			config: Config{
				APIID:                "TEST-12345",
				APIKey:               "test-key-12345",
				DefaultPaymentMethod: MethodDANA,
			},
			expectedPaymentMethod: MethodDANA,
		},
		{
			name: "without default payment method",
			config: Config{
				APIID:  "TEST-12345",
				APIKey: "test-key-12345",
			},
			expectedPaymentMethod: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			if client.defaultPaymentMethod != tt.expectedPaymentMethod {
				t.Errorf("client.defaultPaymentMethod = %v, want %v", client.defaultPaymentMethod, tt.expectedPaymentMethod)
			}
			if client.GetDefaultPaymentMethod() != tt.expectedPaymentMethod {
				t.Errorf("GetDefaultPaymentMethod() = %v, want %v", client.GetDefaultPaymentMethod(), tt.expectedPaymentMethod)
			}
		})
	}
}

// TestGenerateSignature tests signature generation
func TestGenerateSignature(t *testing.T) {
	config := Config{
		APIID:  "SANBOX-90976113",
		APIKey: "SANBOX-snuNYFCZ9q7KhDPpWSTv7243YRUCSrC",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	tests := []struct {
		name        string
		method      string
		merchantRef string
		amount      int64
		expectedSig string
	}{
		{
			name:        "QRIS payment",
			method:      "QRIS",
			merchantRef: "d83heuie230948",
			amount:      20000,
			expectedSig: "",
		},
		{
			name:        "BCAVA payment",
			method:      "BCAVA",
			merchantRef: "d834Hd-dj83kdk-38479freHFH3s",
			amount:      20000,
			expectedSig: "",
		},
		{
			name:        "DANA payment",
			method:      "DANA",
			merchantRef: "test-ref-123",
			amount:      50000,
			expectedSig: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := client.GenerateSignature(tt.method, tt.merchantRef, tt.amount)

			// Signature should be a 64-character hex string (SHA256 = 256 bits = 64 hex chars)
			if len(sig) != 64 {
				t.Errorf("GenerateSignature() returned length %d, want 64", len(sig))
			}

			// Signature should be lowercase hex
			for _, c := range sig {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("GenerateSignature() returned invalid hex character: %c", c)
				}
			}

			// Same inputs should produce same signature
			sig2 := client.GenerateSignature(tt.method, tt.merchantRef, tt.amount)
			if sig != sig2 {
				t.Error("GenerateSignature() returned different signatures for same input")
			}

			// Different inputs should produce different signatures
			differentSig := client.GenerateSignature(tt.method, tt.merchantRef+"-different", tt.amount)
			if sig == differentSig {
				t.Error("GenerateSignature() returned same signature for different input")
			}
		})
	}
}

// TestSignatureConsistency tests signature consistency across different clients with same credentials
func TestSignatureConsistency(t *testing.T) {
	config := Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-67890",
	}

	client1, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	client2, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	method := "QRIS"
	merchantRef := "test-ref"
	amount := int64(10000)

	sig1 := client1.GenerateSignature(method, merchantRef, amount)
	sig2 := client2.GenerateSignature(method, merchantRef, amount)

	if sig1 != sig2 {
		t.Error("GenerateSignature() should produce same signature for same credentials")
	}
}

// TestVerifyCallbackSignature tests callback signature verification
func TestVerifyCallbackSignature(t *testing.T) {
	config := Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-67890",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Create a valid signature
	jsonPayload := []byte(`{"trx_id":"TEST123","merchant_ref":"REF123","status":"berhasil","status_kode":1}`)

	// Generate the correct signature
	h := hmac.New(sha256.New, []byte(client.apiKey))
	h.Write(jsonPayload)
	validSignature := hex.EncodeToString(h.Sum(nil))

	tests := []struct {
		name      string
		payload   []byte
		signature string
		wantValid bool
	}{
		{
			name:      "valid signature",
			payload:   jsonPayload,
			signature: validSignature,
			wantValid: true,
		},
		{
			name:      "invalid signature",
			payload:   jsonPayload,
			signature: "invalid" + validSignature,
			wantValid: false,
		},
		{
			name:      "empty signature",
			payload:   jsonPayload,
			signature: "",
			wantValid: false,
		},
		{
			name:      "different payload",
			payload:   []byte(`{"different":"payload"}`),
			signature: validSignature,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.VerifyCallbackSignature(tt.payload, tt.signature)
			if result != tt.wantValid {
				t.Errorf("VerifyCallbackSignature() = %v, want %v", result, tt.wantValid)
			}
		})
	}
}

// TestBaseURLs tests that the correct base URLs are used
func TestBaseURLs(t *testing.T) {
	tests := []struct {
		name      string
		isSandbox bool
		wantURL   string
	}{
		{
			name:      "production URL",
			isSandbox: false,
			wantURL:   ProductionBaseURL,
		},
		{
			name:      "sandbox URL",
			isSandbox: true,
			wantURL:   SandboxBaseURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(Config{
				APIID:     "TEST-12345",
				APIKey:    "test-key-12345",
				IsSandbox: tt.isSandbox,
			})
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			if client.baseURL != tt.wantURL {
				t.Errorf("baseURL = %v, want %v", client.baseURL, tt.wantURL)
			}

			if client.IsSandbox() != tt.isSandbox {
				t.Errorf("IsSandbox() = %v, want %v", client.IsSandbox(), tt.isSandbox)
			}
		})
	}
}

// TestCustomHTTPClient tests using a custom HTTP client
func TestCustomHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 45 * time.Second,
	}

	config := Config{
		APIID:      "TEST-12345",
		APIKey:     "test-key-12345",
		HTTPClient: customClient,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.httpClient != customClient {
		t.Error("Custom HTTP client was not set")
	}
}

// TestErrorDefinitions tests that error variables are properly defined
func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrMissingAPIID",
			err:  ErrMissingAPIID,
			want: "API ID is required",
		},
		{
			name: "ErrMissingAPIKey",
			err:  ErrMissingAPIKey,
			want: "API Key is required",
		},
		{
			name: "ErrInvalidAmount",
			err:  ErrInvalidAmount,
			want: "invalid amount",
		},
		{
			name: "ErrInvalidPhone",
			err:  ErrInvalidPhone,
			want: "invalid phone number",
		},
		{
			name: "ErrMissingMerchantRef",
			err:  ErrMissingMerchantRef,
			want: "merchant reference is required",
		},
		{
			name: "ErrMissingMethod",
			err:  ErrMissingMethod,
			want: "payment method is required",
		},
		{
			name: "ErrInvalidSignature",
			err:  ErrInvalidSignature,
			want: "invalid signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("Error() = %v, want %v", tt.err.Error(), tt.want)
			}
		})
	}
}

// TestAPIError tests the APIError type
func TestAPIError(t *testing.T) {
	tests := []struct {
		name       string
		err        *APIError
		wantString string
	}{
		{
			name: "error with status",
			err: &APIError{
				StatusCode: 400,
				Status:     "400",
				Message:    "Invalid request",
			},
			wantString: "API error (status 400): Invalid request",
		},
		{
			name: "error without status",
			err: &APIError{
				StatusCode: 500,
				Message:    "Internal server error",
			},
			wantString: "API error (code 500): Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantString {
				t.Errorf("Error() = %v, want %v", tt.err.Error(), tt.wantString)
			}
		})
	}
}
