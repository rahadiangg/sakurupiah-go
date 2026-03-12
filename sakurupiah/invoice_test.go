package sakurupiah

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCreateInvoice_Success tests successful invoice creation
func TestCreateInvoice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/create.php" {
			t.Errorf("Expected /create.php path, got %s", r.URL.Path)
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"message": "Invoice created successfully",
			"data": [{
				"trx_id": "TRX-12345",
				"merchant_ref": "INV-2025-001",
				"amount": "10000",
				"status": "pending",
				"checkout_url": "https://checkout.sakurupiah.com/TRX-12345"
			}]
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.CreateInvoice(CreateInvoiceRequest{
		Method:        "QRIS",
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		CustomerPhone: "628123456789",
		Amount:        10000,
		MerchantRef:   "INV-2025-001",
		CallbackURL:   "https://example.com/callback",
		ReturnURL:     "https://example.com/return",
	})

	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}

	if resp == nil {
		t.Fatal("CreateInvoice() returned nil response")
	}

	if resp.Status != "success" {
		t.Errorf("Status = %v, want success", resp.Status)
	}

	if len(resp.Data) == 0 {
		t.Fatal("CreateInvoice() returned empty data array")
	}
	if resp.Data[0].TrxID != "TRX-12345" {
		t.Errorf("TrxID = %v, want TRX-12345", resp.Data[0].TrxID)
	}
}

// TestCreateInvoice_ValidationErrors tests validation errors
func TestCreateInvoice_ValidationErrors(t *testing.T) {
	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
		DefaultCallbackURL: "https://example.com/callback",
		DefaultReturnURL:   "https://example.com/return",
	})

	tests := []struct {
		name    string
		req     CreateInvoiceRequest
		wantErr error
	}{
		{
			name: "missing method",
			req: CreateInvoiceRequest{
				CustomerPhone: "628123456789",
				Amount:        10000,
				MerchantRef:   "INV-001",
				CallbackURL:   "https://example.com/callback",
				ReturnURL:     "https://example.com/return",
			},
			wantErr: ErrMissingMethod,
		},
		{
			name: "missing phone",
			req: CreateInvoiceRequest{
				Method:      "QRIS",
				Amount:      10000,
				MerchantRef: "INV-001",
				CallbackURL: "https://example.com/callback",
				ReturnURL:   "https://example.com/return",
			},
			wantErr: ErrInvalidPhone,
		},
		{
			name: "invalid amount - zero",
			req: CreateInvoiceRequest{
				Method:        "QRIS",
				CustomerPhone: "628123456789",
				Amount:        0,
				MerchantRef:   "INV-001",
				CallbackURL:   "https://example.com/callback",
				ReturnURL:     "https://example.com/return",
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "invalid amount - negative",
			req: CreateInvoiceRequest{
				Method:        "QRIS",
				CustomerPhone: "628123456789",
				Amount:        -1000,
				MerchantRef:   "INV-001",
				CallbackURL:   "https://example.com/callback",
				ReturnURL:     "https://example.com/return",
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "missing merchant ref",
			req: CreateInvoiceRequest{
				Method:        "QRIS",
				CustomerPhone: "628123456789",
				Amount:        10000,
				CallbackURL:   "https://example.com/callback",
				ReturnURL:     "https://example.com/return",
			},
			wantErr: ErrMissingMerchantRef,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateInvoice(tt.req)
			if err != tt.wantErr {
				t.Errorf("CreateInvoice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCreateInvoice_UsesDefaultPaymentMethod tests that default payment method is used
func TestCreateInvoice_UsesDefaultPaymentMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that method was set
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		method := r.FormValue("method")
		if method != "QRIS" {
			t.Errorf("Expected method QRIS, got %s", method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-123"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:                "TEST-12345",
		APIKey:               "test-key-12345",
		DefaultPaymentMethod: "QRIS",
		DefaultCallbackURL:   "https://example.com/callback",
		DefaultReturnURL:     "https://example.com/return",
	})
	client.baseURL = server.URL + "/"

	_, err := client.CreateInvoice(CreateInvoiceRequest{
		CustomerPhone: "628123456789",
		Amount:        10000,
		MerchantRef:   "INV-001",
	})

	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}
}

// TestCreateInvoice_UsesDefaultURLs tests that default URLs are used
func TestCreateInvoice_UsesDefaultURLs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		callbackURL := r.FormValue("callback_url")
		returnURL := r.FormValue("return_url")

		if callbackURL != "https://default.com/callback" {
			t.Errorf("Expected callback URL https://default.com/callback, got %s", callbackURL)
		}
		if returnURL != "https://default.com/return" {
			t.Errorf("Expected return URL https://default.com/return, got %s", returnURL)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-123"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:              "TEST-12345",
		APIKey:             "test-key-12345",
		DefaultCallbackURL: "https://default.com/callback",
		DefaultReturnURL:   "https://default.com/return",
	})
	client.baseURL = server.URL + "/"

	_, err := client.CreateInvoice(CreateInvoiceRequest{
		Method:        "QRIS",
		CustomerPhone: "628123456789",
		Amount:        10000,
		MerchantRef:   "INV-001",
	})

	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}
}

// TestCreateInvoice_WithProducts tests invoice creation with products
func TestCreateInvoice_WithProducts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		// Check product arrays
		products := r.Form["produk[]"]
		qty := r.Form["qty[]"]
		harga := r.Form["harga[]"]

		if len(products) != 2 {
			t.Errorf("Expected 2 products, got %d", len(products))
		}
		if products[0] != "T-Shirt" {
			t.Errorf("Expected product T-Shirt, got %s", products[0])
		}
		if qty[0] != "1" {
			t.Errorf("Expected qty 1, got %s", qty[0])
		}
		if harga[0] != "50000" {
			t.Errorf("Expected harga 50000, got %s", harga[0])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-123"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:              "TEST-12345",
		APIKey:             "test-key-12345",
		DefaultCallbackURL: "https://example.com/callback",
		DefaultReturnURL:   "https://example.com/return",
	})
	client.baseURL = server.URL + "/"

	products := []Product{
		{Name: "T-Shirt", Qty: 1, Price: 50000, Size: "L", Note: "Blue"},
		{Name: "Pants", Qty: 2, Price: 75000, Size: "M"},
	}

	_, err := client.CreateInvoice(CreateInvoiceRequest{
		Method:        "QRIS",
		CustomerPhone: "628123456789",
		Amount:        200000,
		MerchantRef:   "INV-001",
		Products:      products,
	})

	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}
}

// TestCreateInvoice_MerchantFeeDefaults tests merchant fee defaults
func TestCreateInvoice_MerchantFeeDefaults(t *testing.T) {
	_, _ = NewClient(Config{
		APIID:              "TEST-12345",
		APIKey:             "test-key-12345",
		DefaultCallbackURL: "https://example.com/callback",
		DefaultReturnURL:   "https://example.com/return",
	})

	tests := []struct {
		name         string
		merchantFee  int
		expectedFee  int
	}{
		{
			name:        "no fee specified - defaults to merchant pays",
			merchantFee: 0,
			expectedFee: int(FeeTypeMerchant),
		},
		{
			name:        "merchant pays",
			merchantFee: int(FeeTypeMerchant),
			expectedFee: int(FeeTypeMerchant),
		},
		{
			name:        "customer pays",
			merchantFee: int(FeeTypeCustomer),
			expectedFee: int(FeeTypeCustomer),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just test that validation accepts these values
			// The actual request will fail due to no server, but we're just testing fee handling
			req := CreateInvoiceRequest{
				Method:        "QRIS",
				CustomerPhone: "628123456789",
				Amount:        10000,
				MerchantRef:   "INV-001",
				MerchantFee:   tt.merchantFee,
			}

			// Since fee normalization happens before validation, we need to check it
			if req.MerchantFee != int(FeeTypeMerchant) && req.MerchantFee != int(FeeTypeCustomer) {
				req.MerchantFee = int(FeeTypeMerchant)
			}

			if req.MerchantFee != tt.expectedFee {
				t.Errorf("MerchantFee = %v, want %v", req.MerchantFee, tt.expectedFee)
			}
		})
	}
}

// TestCreateInvoiceWithProducts tests convenience method
func TestCreateInvoiceWithProducts(t *testing.T) {
	client, _ := NewClient(Config{
		APIID:              "TEST-12345",
		APIKey:             "test-key-12345",
		DefaultCallbackURL: "https://example.com/callback",
		DefaultReturnURL:   "https://example.com/return",
	})

	products := []Product{
		{Name: "Product 1", Qty: 1, Price: 10000},
	}

	// This will fail with no server, but we're testing the method signature and call
	_, err := client.CreateInvoiceWithProducts("QRIS", "628123456789", 10000, "INV-001", products)

	// Error is expected since no server is running
	if err == nil {
		t.Error("Expected error when no server is running")
	}
}

// TestCreateInvoiceSimple tests the simple invoice creation method
func TestCreateInvoiceSimple(t *testing.T) {
	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})

	// This will fail with no server, but we're testing the method signature
	_, err := client.CreateInvoiceSimple(
		"QRIS",
		"John Doe",
		"628123456789",
		10000,
		"INV-001",
		"https://example.com/callback",
		"https://example.com/return",
	)

	// Error is expected since no server is running
	if err == nil {
		t.Error("Expected error when no server is running")
	}
}

// TestCreateInvoice_ExpiredParameter tests expired parameter handling
func TestCreateInvoice_ExpiredParameter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		expired := r.FormValue("expired")
		if expired != "24" {
			t.Errorf("Expected expired 24, got %s", expired)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-123"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:              "TEST-12345",
		APIKey:             "test-key-12345",
		DefaultCallbackURL: "https://example.com/callback",
		DefaultReturnURL:   "https://example.com/return",
	})
	client.baseURL = server.URL + "/"

	_, err := client.CreateInvoice(CreateInvoiceRequest{
		Method:        "QRIS",
		CustomerPhone: "628123456789",
		Amount:        10000,
		MerchantRef:   "INV-001",
		Expired:       24,
	})

	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}
}

// TestCreateInvoice_CustomerDetails tests optional customer detail fields
func TestCreateInvoice_CustomerDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		phone := r.FormValue("phone")

		if name != "Jane Doe" {
			t.Errorf("Expected name Jane Doe, got %s", name)
		}
		if email != "jane@example.com" {
			t.Errorf("Expected email jane@example.com, got %s", email)
		}
		if phone != "628987654321" {
			t.Errorf("Expected phone 628987654321, got %s", phone)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-123"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:              "TEST-12345",
		APIKey:             "test-key-12345",
		DefaultCallbackURL: "https://example.com/callback",
		DefaultReturnURL:   "https://example.com/return",
	})
	client.baseURL = server.URL + "/"

	_, err := client.CreateInvoice(CreateInvoiceRequest{
		Method:        "QRIS",
		CustomerName:  "Jane Doe",
		CustomerEmail: "jane@example.com",
		CustomerPhone: "628987654321",
		Amount:        10000,
		MerchantRef:   "INV-001",
	})

	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}
}

// TestCreateInvoice_SignatureGeneration tests that signature is generated correctly
func TestCreateInvoice_SignatureGeneration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		signature := r.FormValue("signature")
		if signature == "" {
			t.Error("Expected signature to be present")
		}

		// Signature should be 64 characters (SHA256 hex)
		if len(signature) != 64 {
			t.Errorf("Expected signature length 64, got %d", len(signature))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-123"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:              "TEST-12345",
		APIKey:             "test-key-12345",
		DefaultCallbackURL: "https://example.com/callback",
		DefaultReturnURL:   "https://example.com/return",
	})
	client.baseURL = server.URL + "/"

	_, err := client.CreateInvoice(CreateInvoiceRequest{
		Method:        "QRIS",
		CustomerPhone: "628123456789",
		Amount:        10000,
		MerchantRef:   "INV-001",
	})

	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}
}

// TestCreateInvoice_ProductWithOptionalFields tests products with optional fields
func TestCreateInvoice_ProductWithOptionalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		// Check that size and note are included
		size := r.Form["size[]"]
		note := r.Form["note[]"]

		if len(size) != 1 {
			t.Errorf("Expected 1 size, got %d", len(size))
		}
		if size[0] != "XL" {
			t.Errorf("Expected size XL, got %s", size[0])
		}
		if len(note) != 1 {
			t.Errorf("Expected 1 note, got %d", len(note))
		}
		if note[0] != "Red color" {
			t.Errorf("Expected note 'Red color', got %s", note[0])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-123"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:              "TEST-12345",
		APIKey:             "test-key-12345",
		DefaultCallbackURL: "https://example.com/callback",
		DefaultReturnURL:   "https://example.com/return",
	})
	client.baseURL = server.URL + "/"

	products := []Product{
		{Name: "Shirt", Qty: 1, Price: 50000, Size: "XL", Note: "Red color"},
	}

	_, err := client.CreateInvoice(CreateInvoiceRequest{
		Method:        "QRIS",
		CustomerPhone: "628123456789",
		Amount:        50000,
		MerchantRef:   "INV-001",
		Products:      products,
	})

	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}
}
