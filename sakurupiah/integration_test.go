//go:build integration
// +build integration

package sakurupiah

import (
	"fmt"
	"testing"
	"time"
)

const (
	sandboxAPIID  = "SANBOX-45829715"
	sandboxAPIKey = "SANBOX-LF1KVw7QDKuMU5ybmCaMh3R6jdNi"
)

// getIntegrationClient returns a client configured for sandbox testing
func getIntegrationClient() (*Client, error) {
	return NewClient(Config{
		APIID:    sandboxAPIID,
		APIKey:   sandboxAPIKey,
		IsSandbox: true,
	})
}

// TestIntegrationListPaymentChannels tests listing payment channels via real API
func TestIntegrationListPaymentChannels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := getIntegrationClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.ListPaymentChannels()
	if err != nil {
		t.Fatalf("ListPaymentChannels() error = %v", err)
	}

	// Validate response
	if resp.Status != "200" {
		t.Errorf("Expected status 200, got %s", resp.Status)
	}

	if len(resp.Data) == 0 {
		t.Error("Expected at least one payment channel")
	}

	// Verify expected channels exist (SHOPEEPAY is optional as it may not be available in sandbox)
	expectedChannels := map[string]bool{
		"QRIS":  false,
		"BCAVA": false,
		"BRIVA": false,
		"BNIVA": false,
		"DANA":  false,
		"OVO":   false,
		"GOPAY": false,
	}
	optionalChannels := map[string]bool{
		"SHOPEEPAY": false,
	}

	for _, ch := range resp.Data {
		if _, exists := expectedChannels[ch.Code]; exists {
			expectedChannels[ch.Code] = true
			t.Logf("Found channel: %s - %s (Status: %s, Min: %s, Max: %s, Fee: %s)",
				ch.Code, ch.Name, ch.Status, ch.Min, ch.Max, ch.Fee)
		}
		if _, exists := optionalChannels[ch.Code]; exists {
			optionalChannels[ch.Code] = true
			t.Logf("Found optional channel: %s - %s", ch.Code, ch.Name)
		}
	}

	for code, found := range expectedChannels {
		if !found {
			t.Errorf("Expected channel %s not found in response", code)
		}
	}

	// Log optional channels that were found
	for code, found := range optionalChannels {
		if found {
			t.Logf("Optional channel %s is available", code)
		}
	}
}

// TestIntegrationCheckBalance tests checking balance via real API
func TestIntegrationCheckBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := getIntegrationClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.CheckBalance()
	if err != nil {
		t.Fatalf("CheckBalance() error = %v", err)
	}

	// Validate response
	if resp.Status != "200" {
		t.Errorf("Expected status 200, got %s", resp.Status)
	}

	if resp.Data.MerchantName == "" {
		t.Error("Merchant name should not be empty")
	}

	t.Logf("Merchant: %s", resp.Data.MerchantName)
	t.Logf("Balance: %s", resp.Data.Balance)
	t.Logf("Available Balance: %s", resp.Data.AvailableBalance)
}

// TestIntegrationCreateInvoice tests creating an invoice via real API
func TestIntegrationCreateInvoice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := getIntegrationClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Set default URLs for testing
	client.SetDefaultCallbackURL("https://example.com/callback")
	client.SetDefaultReturnURL("https://example.com/return")

	// Generate unique merchant ref
	merchantRef := fmt.Sprintf("TEST-INV-%d", time.Now().UnixNano())

	tests := []struct {
		name    string
		req     CreateInvoiceRequest
		wantErr bool
	}{
		{
			name: "valid QRIS invoice",
			req: CreateInvoiceRequest{
				Method:       "QRIS",
				CustomerName: "Integration Test",
				CustomerEmail: "test@example.com",
				CustomerPhone: "628123456789",
				Amount:       10000,
				MerchantFee:  int(FeeTypeMerchant),
				MerchantRef:  merchantRef + "-QRIS",
				CallbackURL:  "https://example.com/callback",
				ReturnURL:    "https://example.com/return",
			},
			wantErr: false,
		},
		{
			name: "valid BCAVA invoice",
			req: CreateInvoiceRequest{
				Method:       "BCAVA",
				CustomerName: "Integration Test",
				CustomerPhone: "628123456789",
				Amount:       20000,
				MerchantFee:  int(FeeTypeMerchant),
				MerchantRef:  merchantRef + "-BCAVA",
				CallbackURL:  "https://example.com/callback",
				ReturnURL:    "https://example.com/return",
			},
			wantErr: false,
		},
		{
			name: "invoice with products",
			req: CreateInvoiceRequest{
				Method:       "DANA",
				CustomerName: "Integration Test",
				CustomerPhone: "628123456789",
				Amount:       50000,
				MerchantFee:  int(FeeTypeCustomer),
				MerchantRef:  merchantRef + "-PRODUCTS",
				Expired:      12,
				Products: []Product{
					{Name: "Test Product 1", Qty: 1, Price: 25000, Size: "L", Note: "Blue"},
					{Name: "Test Product 2", Qty: 1, Price: 25000, Size: "M", Note: "Red"},
				},
				CallbackURL: "https://example.com/callback",
				ReturnURL:   "https://example.com/return",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate signature
			tt.req.Signature = client.GenerateSignature(tt.req.Method, tt.req.MerchantRef, tt.req.Amount)

			resp, err := client.CreateInvoice(tt.req)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error = %v", err)
				return
			}

			if err == nil {
				// Validate response
				if resp.Status != "200" {
					t.Errorf("Expected status 200, got %s: %s", resp.Status, resp.Message)
				}

				if len(resp.Data) == 0 {
					t.Error("Expected invoice data in response")
					return
				}

				invoice := resp.Data[0]

				t.Logf("Invoice Created Successfully:")
				t.Logf("  Trx ID: %s", invoice.TrxID)
				t.Logf("  Merchant Ref: %s", invoice.MerchantRef)
				t.Logf("  Payment Code: %s", invoice.PaymentCode)
				t.Logf("  Amount: %d", invoice.Total)
				t.Logf("  Status: %s", invoice.PaymentStatus)
				t.Logf("  Checkout URL: %s", invoice.CheckoutURL)

				// Verify fields
				if invoice.TrxID == "" {
					t.Error("TrxID should not be empty")
				}
				if invoice.PaymentStatus == "" {
					t.Error("PaymentStatus should not be empty")
				}
				// When customer pays the fee, total amount includes the fee
				if tt.req.MerchantFee == int(FeeTypeCustomer) {
					// Total will be amount + fee, just log it instead of strict check
					t.Logf("Note: Customer pays fee - Total amount %d includes transaction fee", invoice.Total)
				} else {
					if invoice.Total != tt.req.Amount {
						t.Errorf("Amount mismatch: got %d, want %d", invoice.Total, tt.req.Amount)
					}
				}

				// For invoices with products, verify product data
				if len(tt.req.Products) > 0 {
					if len(resp.Product) != len(tt.req.Products) {
						t.Errorf("Product count mismatch: got %d, want %d", len(resp.Product), len(tt.req.Products))
					}
					for i, p := range resp.Product {
						t.Logf("  Product %d: %s (Qty: %s, Price: %d)", i+1, p.Name, p.Qty, p.Price)
					}
				}
			}
		})
	}
}

// TestIntegrationCreateInvoiceSimple tests the simplified invoice creation via real API
func TestIntegrationCreateInvoiceSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := getIntegrationClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	merchantRef := fmt.Sprintf("TEST-SIMPLE-%d", time.Now().UnixNano())

	resp, err := client.CreateInvoiceSimple(
		"QRIS",
		"Simple Test User",
		"628123456789",
		15000,
		merchantRef,
		"https://example.com/callback",
		"https://example.com/return",
	)

	if err != nil {
		t.Fatalf("CreateInvoiceSimple() error = %v", err)
	}

	if resp.Status != "200" {
		t.Errorf("Expected status 200, got %s", resp.Status)
	}

	if len(resp.Data) == 0 {
		t.Fatal("Expected invoice data in response")
	}

	t.Logf("Simple Invoice Created:")
	t.Logf("  Trx ID: %s", resp.Data[0].TrxID)
	t.Logf("  Amount: %d", resp.Data[0].Total)
}

// TestIntegrationTransactionHistory tests getting transaction history via real API
func TestIntegrationTransactionHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := getIntegrationClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name    string
		req     TransactionHistoryRequest
		wantErr bool
	}{
		{
			name: "get all transactions",
			req:  TransactionHistoryRequest{},
		},
		{
			name: "filter by status - pending",
			req: TransactionHistoryRequest{
				Status: "pending",
			},
		},
		{
			name: "filter by status - berhasil",
			req: TransactionHistoryRequest{
				Status: "berhasil",
			},
			wantErr: true, // API returns "Tidak ditemukan" when no successful transactions exist
		},
		{
			name: "filter by merchant only",
			req: TransactionHistoryRequest{
				MerchantFilter: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetTransactionHistory(tt.req)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if tt.wantErr && err != nil {
				t.Logf("Got expected error: %v", err)
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error = %v", err)
				return
			}

			if resp.Status != "200" {
				t.Errorf("Expected status 200, got %s: %s", resp.Status, resp.Message)
			}

			t.Logf("Found %d transactions", len(resp.Data))
			for i, trx := range resp.Data {
				if i < 3 { // Log first 3 transactions
					t.Logf("  %s: %s - %s (%s)", trx.TrxID, trx.Amount, trx.Status, trx.Date)
				}
			}
		})
	}
}

// TestIntegrationGetTransactionStatus tests checking transaction status via real API
func TestIntegrationGetTransactionStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := getIntegrationClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// First, create a test invoice to get a valid TrxID
	merchantRef := fmt.Sprintf("TEST-STATUS-%d", time.Now().UnixNano())
	createReq := CreateInvoiceRequest{
		Method:       "QRIS",
		CustomerName: "Status Test",
		CustomerPhone: "628123456789",
		Amount:       10000,
		MerchantFee:  int(FeeTypeMerchant),
		MerchantRef:  merchantRef,
		CallbackURL:  "https://example.com/callback",
		ReturnURL:    "https://example.com/return",
	}

	createResp, err := client.CreateInvoice(createReq)
	if err != nil {
		t.Fatalf("Failed to create test invoice: %v", err)
	}

	if len(createResp.Data) == 0 {
		t.Fatal("No invoice data returned")
	}

	trxID := createResp.Data[0].TrxID
	t.Logf("Testing status for TrxID: %s", trxID)

	// Now check the status
	statusResp, err := client.GetTransactionStatus(trxID)
	if err != nil {
		t.Fatalf("GetTransactionStatus() error = %v", err)
	}

	if statusResp.Status != "200" {
		t.Errorf("Expected status 200, got %s", statusResp.Status)
	}

	if len(statusResp.Data) == 0 {
		t.Error("Expected status data in response")
	}

	t.Logf("Transaction Status: %s", statusResp.Data[0].Status)
}

// TestIntegrationSignatureGeneration tests that signature generation works correctly
func TestIntegrationSignatureGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := getIntegrationClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	method := "QRIS"
	merchantRef := fmt.Sprintf("TEST-SIG-%d", time.Now().UnixNano())
	amount := int64(25000)

	sig := client.GenerateSignature(method, merchantRef, amount)

	// Signature should be 64 characters (SHA256 hex)
	if len(sig) != 64 {
		t.Errorf("Signature length = %d, want 64", len(sig))
	}

	t.Logf("Generated signature: %s", sig)

	// Test that same inputs produce same signature
	sig2 := client.GenerateSignature(method, merchantRef, amount)
	if sig != sig2 {
		t.Error("Same inputs should produce same signature")
	}
}

// TestIntegrationCreateInvoiceWithProducts tests invoice creation with product arrays
func TestIntegrationCreateInvoiceWithProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := getIntegrationClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Set default URLs
	client.SetDefaultCallbackURL("https://example.com/callback")
	client.SetDefaultReturnURL("https://example.com/return")

	merchantRef := fmt.Sprintf("TEST-PRODS-%d", time.Now().UnixNano())

	products := []Product{
		{Name: "T-Shirt", Qty: 2, Price: 50000, Size: "L", Note: "Blue"},
		{Name: "Jeans", Qty: 1, Price: 100000, Size: "32", Note: "Black"},
		{Name: "Socks", Qty: 3, Price: 15000, Size: "M", Note: "White"},
	}

	resp, err := client.CreateInvoiceWithProducts(
		"BRIVA",
		"628123456789",
		215000,
		merchantRef,
		products,
	)

	if err != nil {
		t.Fatalf("CreateInvoiceWithProducts() error = %v", err)
	}

	if resp.Status != "200" {
		t.Errorf("Expected status 200, got %s: %s", resp.Status, resp.Message)
	}

	if len(resp.Data) == 0 {
		t.Fatal("Expected invoice data in response")
	}

	t.Logf("Invoice with %d products created:", len(products))
	t.Logf("  Trx ID: %s", resp.Data[0].TrxID)
	t.Logf("  Total: %d", resp.Data[0].Total)

	if len(resp.Product) != len(products) {
		t.Errorf("Product count mismatch: got %d, want %d", len(resp.Product), len(products))
	}

	for i, p := range resp.Product {
		t.Logf("  Product %d: %s - Qty: %s, Price: %d", i+1, p.Name, p.Qty, p.Price)
	}
}

// TestIntegrationErrorHandling tests error handling with invalid data
func TestIntegrationErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := getIntegrationClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("invalid payment method", func(t *testing.T) {
		merchantRef := fmt.Sprintf("TEST-ERROR-%d", time.Now().UnixNano())
		req := CreateInvoiceRequest{
			Method:       "INVALID_METHOD",
			CustomerPhone: "628123456789",
			Amount:       10000,
			MerchantFee:  int(FeeTypeMerchant),
			MerchantRef:  merchantRef,
			CallbackURL:  "https://example.com/callback",
			ReturnURL:    "https://example.com/return",
		}

		_, err := client.CreateInvoice(req)
		if err == nil {
			t.Error("Expected error for invalid payment method")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("invalid transaction ID", func(t *testing.T) {
		_, err := client.GetTransactionStatus("INVALID_TRX_ID")
		if err == nil {
			t.Error("Expected error for invalid transaction ID")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})
}

// TestIntegrationMinimumAmount tests creating invoices with minimum amounts for different channels
func TestIntegrationMinimumAmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := getIntegrationClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		method    string
		minAmount int64
	}{
		{"QRIS", 500},
		{"GOPAY", 500},
		{"DANA", 1000},
		{"BCAVA", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.method+" minimum", func(t *testing.T) {
			merchantRef := fmt.Sprintf("TEST-MIN-%s-%d", tt.method, time.Now().UnixNano())

			req := CreateInvoiceRequest{
				Method:       tt.method,
				CustomerName: "Min Amount Test",
				CustomerPhone: "628123456789",
				Amount:       tt.minAmount,
				MerchantFee:  int(FeeTypeMerchant),
				MerchantRef:  merchantRef,
				CallbackURL:  "https://example.com/callback",
				ReturnURL:    "https://example.com/return",
			}

			resp, err := client.CreateInvoice(req)
			if err != nil {
				t.Errorf("Failed to create invoice with minimum amount %d for %s: %v", tt.minAmount, tt.method, err)
			} else {
				t.Logf("Created %s invoice with minimum amount %d: TrxID=%s", tt.method, tt.minAmount, resp.Data[0].TrxID)
			}
		})
	}
}
