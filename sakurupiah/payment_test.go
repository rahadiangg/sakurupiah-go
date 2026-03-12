package sakurupiah

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestListPaymentChannels_Success tests successful payment channels list
func TestListPaymentChannels_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/list-payment.php" {
			t.Errorf("Expected /list-payment.php path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"message": "Payment channels retrieved",
			"data": [
				{
					"kode": "QRIS",
					"nama": "QRIS",
					"status": "Aktif",
					"logo": "https://example.com/qris.png",
					"biaya": "0",
					"percent": "0.7",
					"minimal": "100",
					"maksimal": "100000000",
					"tipe": "DIRECT"
				},
				{
					"kode": "BCAVA",
					"nama": "BCA Virtual Account",
					"status": "Aktif",
					"logo": "https://example.com/bca.png",
					"biaya": "4000",
					"percent": "0",
					"minimal": "10000",
					"maksimal": "100000000",
					"tipe": "DIRECT"
				}
			]
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.ListPaymentChannels()

	if err != nil {
		t.Fatalf("ListPaymentChannels() error = %v", err)
	}

	if resp == nil {
		t.Fatal("ListPaymentChannels() returned nil response")
	}

	if resp.Status != "success" {
		t.Errorf("Status = %v, want success", resp.Status)
	}

	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(resp.Data))
	}

	// Check first channel
	qris := resp.Data[0]
	if qris.Code != "QRIS" {
		t.Errorf("First channel code = %v, want QRIS", qris.Code)
	}
	if qris.Status != "Aktif" {
		t.Errorf("First channel status = %v, want Aktif", qris.Status)
	}
	if qris.Type != "DIRECT" {
		t.Errorf("First channel type = %v, want DIRECT", qris.Type)
	}

	// Check second channel
	bca := resp.Data[1]
	if bca.Code != "BCAVA" {
		t.Errorf("Second channel code = %v, want BCAVA", bca.Code)
	}
	if bca.Name != "BCA Virtual Account" {
		t.Errorf("Second channel name = %v, want BCA Virtual Account", bca.Name)
	}
}

// TestListPaymentChannels_EmptyList tests handling of empty channel list
func TestListPaymentChannels_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "No payment channels available",
			"data": []
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.ListPaymentChannels()

	if err != nil {
		t.Fatalf("ListPaymentChannels() error = %v", err)
	}

	if len(resp.Data) != 0 {
		t.Errorf("Expected 0 channels, got %d", len(resp.Data))
	}
}

// TestListPaymentChannels_InactiveChannels tests response with inactive channels
func TestListPaymentChannels_InactiveChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "Payment channels retrieved",
			"data": [
				{
					"kode": "QRIS",
					"nama": "QRIS",
					"status": "Aktif",
					"biaya": "0",
					"percent": "0.7",
					"tipe": "DIRECT"
				},
				{
					"kode": "ALFA",
					"nama": "Alfamart",
					"status": "Offline",
					"biaya": "5000",
					"percent": "0",
					"tipe": "REDIRECT"
				}
			]
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.ListPaymentChannels()

	if err != nil {
		t.Fatalf("ListPaymentChannels() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("Expected 2 channels, got %d", len(resp.Data))
	}

	// Verify we can distinguish active from inactive
	activeCount := 0
	for _, ch := range resp.Data {
		if ch.Status == "Aktif" {
			activeCount++
		}
	}

	if activeCount != 1 {
		t.Errorf("Expected 1 active channel, got %d", activeCount)
	}
}

// TestListPaymentChannels_WithGuide tests channels with payment guide
func TestListPaymentChannels_WithGuide(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "Payment channels retrieved",
			"data": [
				{
					"kode": "QRIS",
					"nama": "QRIS",
					"status": "Aktif",
					"biaya": "0",
					"percent": "0.7",
					"tipe": "DIRECT",
					"guide": {
						"title": "QRIS Payment",
						"payment_guide": "Scan QR code with your e-wallet app"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.ListPaymentChannels()

	if err != nil {
		t.Fatalf("ListPaymentChannels() error = %v", err)
	}

	if len(resp.Data) == 0 {
		t.Fatal("Expected at least 1 channel")
	}

	if resp.Data[0].Guide.PaymentGuide != "Scan QR code with your e-wallet app" {
		t.Errorf("Guide = %v, want 'Scan QR code with your e-wallet app'", resp.Data[0].Guide.PaymentGuide)
	}
}

// TestListPaymentChannels_RequestParams tests that correct request parameters are sent
func TestListPaymentChannels_RequestParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		apiID := r.FormValue("api_id")
		method := r.FormValue("method")

		if apiID != "TEST-12345" {
			t.Errorf("Expected api_id TEST-12345, got %s", apiID)
		}
		if method != "list" {
			t.Errorf("Expected method 'list', got %s", method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	_, err := client.ListPaymentChannels()

	if err != nil {
		t.Fatalf("ListPaymentChannels() error = %v", err)
	}
}

// TestCheckBalance_Success tests successful balance check
func TestCheckBalance_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/check_balance.php" {
			t.Errorf("Expected /check_balance.php path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"message": "Balance retrieved",
			"data": {
				"nama_merchant": "Test Store",
				"balance": "1500000",
				"saldo_tersedia": "1000000"
			}
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.CheckBalance()

	if err != nil {
		t.Fatalf("CheckBalance() error = %v", err)
	}

	if resp == nil {
		t.Fatal("CheckBalance() returned nil response")
	}

	if resp.Status != "success" {
		t.Errorf("Status = %v, want success", resp.Status)
	}

	if resp.Data.MerchantName != "Test Store" {
		t.Errorf("MerchantName = %v, want Test Store", resp.Data.MerchantName)
	}

	if resp.Data.Balance != "1500000" {
		t.Errorf("Balance = %v, want 1500000", resp.Data.Balance)
	}

	if resp.Data.AvailableBalance != "1000000" {
		t.Errorf("AvailableBalance = %v, want 1000000", resp.Data.AvailableBalance)
	}
}

// TestCheckBalance_ZeroBalance tests zero balance handling
func TestCheckBalance_ZeroBalance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "Balance retrieved",
			"data": {
				"nama_merchant": "New Store",
				"balance": "0",
				"saldo_tersedia": "0"
			}
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.CheckBalance()

	if err != nil {
		t.Fatalf("CheckBalance() error = %v", err)
	}

	if resp.Data.Balance != "0" {
		t.Errorf("Balance = %v, want 0", resp.Data.Balance)
	}

	if resp.Data.AvailableBalance != "0" {
		t.Errorf("AvailableBalance = %v, want 0", resp.Data.AvailableBalance)
	}
}

// TestCheckBalance_PendingSettlement tests balance with pending settlement
func TestCheckBalance_PendingSettlement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "Balance retrieved",
			"data": {
				"nama_merchant": "Busy Store",
				"balance": "5000000",
				"saldo_tersedia": "1000000"
			}
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.CheckBalance()

	if err != nil {
		t.Fatalf("CheckBalance() error = %v", err)
	}

	// This scenario shows pending settlement (4M pending, 1M available)
	if resp.Data.Balance != "5000000" {
		t.Errorf("Balance = %v, want 5000000", resp.Data.Balance)
	}
}

// TestCheckBalance_RequestParams tests that correct request parameters are sent
func TestCheckBalance_RequestParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		apiID := r.FormValue("api_id")
		method := r.FormValue("method")

		if apiID != "TEST-12345" {
			t.Errorf("Expected api_id TEST-12345, got %s", apiID)
		}
		if method != "balance" {
			t.Errorf("Expected method 'balance', got %s", method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":{"nama_merchant":"Test","balance":"0","saldo_tersedia":"0"}}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	_, err := client.CheckBalance()

	if err != nil {
		t.Fatalf("CheckBalance() error = %v", err)
	}
}

// TestCheckBalance_APIErrorResponse tests API error handling
func TestCheckBalance_APIErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{
			"status": "error",
			"message": "Invalid API credentials"
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	_, err := client.CheckBalance()

	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
}

// TestListPaymentChannels_APIErrorResponse tests error handling for payment channels
func TestListPaymentChannels_APIErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{
			"status": "error",
			"message": "Internal server error"
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	_, err := client.ListPaymentChannels()

	if err == nil {
		t.Error("Expected error for server error")
	}
}

// TestCheckBalance_WithSpecialCharacters tests merchant name with special characters
func TestCheckBalance_WithSpecialCharacters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "Balance retrieved",
			"data": {
				"nama_merchant": "Toko Serba Ada & Co.",
				"balance": "2000000",
				"saldo_tersedia": "1500000"
			}
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.CheckBalance()

	if err != nil {
		t.Fatalf("CheckBalance() error = %v", err)
	}

	if resp.Data.MerchantName != "Toko Serba Ada & Co." {
		t.Errorf("MerchantName = %v, want 'Toko Serba Ada & Co.'", resp.Data.MerchantName)
	}
}

// TestListPaymentChannels_WithFeeStructure tests different fee structures
func TestListPaymentChannels_WithFeeStructure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "Payment channels retrieved",
			"data": [
				{
					"kode": "QRIS",
					"nama": "QRIS",
					"status": "Aktif",
					"biaya": "0",
					"percent": "0.7",
					"tipe": "DIRECT"
				},
				{
					"kode": "BCAVA",
					"nama": "BCA VA",
					"status": "Aktif",
					"biaya": "4000",
					"percent": "0",
					"tipe": "DIRECT"
				},
				{
					"kode": "DANA",
					"nama": "DANA",
					"status": "Aktif",
					"biaya": "1000",
					"percent": "0.5",
					"tipe": "REDIRECT"
				}
			]
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.ListPaymentChannels()

	if err != nil {
		t.Fatalf("ListPaymentChannels() error = %v", err)
	}

	if len(resp.Data) != 3 {
		t.Fatalf("Expected 3 channels, got %d", len(resp.Data))
	}

	// Check different fee structures using the string fields directly
	if resp.Data[0].Percent != "0.7" {
		t.Errorf("QRIS fee percent = %v, want 0.7", resp.Data[0].Percent)
	}
	if resp.Data[1].Fee != "4000" {
		t.Errorf("BCAVA fee = %v, want 4000", resp.Data[1].Fee)
	}
	if resp.Data[2].Fee != "1000" || resp.Data[2].Percent != "0.5" {
		t.Errorf("DANA fee = %s percent:%s, want 1000 percent:0.5",
			resp.Data[2].Fee, resp.Data[2].Percent)
	}
}
