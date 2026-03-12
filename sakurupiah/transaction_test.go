package sakurupiah

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetTransactionHistory_Success tests successful transaction history retrieval
func TestGetTransactionHistory_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/transaction.php" {
			t.Errorf("Expected /transaction.php path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"message": "Transactions retrieved",
			"data": [
				{
					"trx_id": "TRX-12345",
					"merchant_ref": "INV-001",
					"status": "berhasil",
					"amount": "10000"
				},
				{
					"trx_id": "TRX-67890",
					"merchant_ref": "INV-002",
					"status": "pending",
					"amount": "20000"
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

	resp, err := client.GetTransactionHistory(TransactionHistoryRequest{})

	if err != nil {
		t.Fatalf("GetTransactionHistory() error = %v", err)
	}

	if resp.Status != "success" {
		t.Errorf("Status = %v, want success", resp.Status)
	}

	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(resp.Data))
	}
}

// TestGetTransactionHistory_WithFilters tests transaction history with filters
func TestGetTransactionHistory_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		// Verify filters are sent
		status := r.FormValue("status")
		startDate := r.FormValue("tanggal_awal")
		endDate := r.FormValue("tanggal_akhir")

		if status != "berhasil" {
			t.Errorf("Expected status 'berhasil', got %s", status)
		}
		if startDate != "2025-01-01" {
			t.Errorf("Expected start date '2025-01-01', got %s", startDate)
		}
		if endDate != "2025-01-31" {
			t.Errorf("Expected end date '2025-01-31', got %s", endDate)
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

	_, err := client.GetTransactionHistory(TransactionHistoryRequest{
		Status:    "berhasil",
		StartDate: "2025-01-01",
		EndDate:   "2025-01-31",
	})

	if err != nil {
		t.Fatalf("GetTransactionHistory() error = %v", err)
	}
}

// TestGetAllTransactions tests getting all transactions
func TestGetAllTransactions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "All transactions retrieved",
			"data": [
				{"trx_id": "TRX-1", "status": "berhasil"},
				{"trx_id": "TRX-2", "status": "pending"}
			]
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.GetAllTransactions()

	if err != nil {
		t.Fatalf("GetAllTransactions() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(resp.Data))
	}
}

// TestGetTransactionsByStatus tests filtering by status
func TestGetTransactionsByStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		status := r.FormValue("status")
		if status != "berhasil" {
			t.Errorf("Expected status 'berhasil', got %s", status)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-1","status":"berhasil"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.GetTransactionsByStatus("berhasil")

	if err != nil {
		t.Fatalf("GetTransactionsByStatus() error = %v", err)
	}

	if len(resp.Data) == 0 {
		t.Error("Expected at least 1 transaction")
	}
}

// TestGetTransactionsByPaymentCode tests filtering by payment code
func TestGetTransactionsByPaymentCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		paymentCode := r.FormValue("payment_kode")
		if paymentCode != "QRIS" {
			t.Errorf("Expected payment_kode 'QRIS', got %s", paymentCode)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-1","payment_kode":"QRIS"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.GetTransactionsByPaymentCode("QRIS")

	if err != nil {
		t.Fatalf("GetTransactionsByPaymentCode() error = %v", err)
	}

	if len(resp.Data) == 0 {
		t.Error("Expected at least 1 transaction")
	}
}

// TestGetTransactionsByMerchantRef tests filtering by merchant reference
func TestGetTransactionsByMerchantRef(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		merchantRef := r.FormValue("merchant_ref")
		if merchantRef != "INV-2025-001" {
			t.Errorf("Expected merchant_ref 'INV-2025-001', got %s", merchantRef)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-1","merchant_ref":"INV-2025-001"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.GetTransactionsByMerchantRef("INV-2025-001")

	if err != nil {
		t.Fatalf("GetTransactionsByMerchantRef() error = %v", err)
	}

	if len(resp.Data) == 0 {
		t.Error("Expected at least 1 transaction")
	}
}

// TestGetTransactionsByDateRange tests filtering by date range
func TestGetTransactionsByDateRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		startDate := r.FormValue("tanggal_awal")
		endDate := r.FormValue("tanggal_akhir")

		if startDate != "2025-01-01" {
			t.Errorf("Expected tanggal_awal '2025-01-01', got %s", startDate)
		}
		if endDate != "2025-01-31" {
			t.Errorf("Expected tanggal_akhir '2025-01-31', got %s", endDate)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-1"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.GetTransactionsByDateRange("2025-01-01", "2025-01-31")

	if err != nil {
		t.Fatalf("GetTransactionsByDateRange() error = %v", err)
	}

	if len(resp.Data) == 0 {
		t.Error("Expected at least 1 transaction")
	}
}

// TestGetTransactionByTrxID tests getting transaction by ID
func TestGetTransactionByTrxID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		trxID := r.FormValue("trx_id")
		if trxID != "TRX-12345" {
			t.Errorf("Expected trx_id 'TRX-12345', got %s", trxID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"OK","data":[{"trx_id":"TRX-12345","status":"berhasil"}]}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.GetTransactionByTrxID("TRX-12345")

	if err != nil {
		t.Fatalf("GetTransactionByTrxID() error = %v", err)
	}

	if len(resp.Data) == 0 {
		t.Fatal("Expected at least 1 transaction")
	}

	if resp.Data[0].TrxID != "TRX-12345" {
		t.Errorf("TrxID = %v, want TRX-12345", resp.Data[0].TrxID)
	}
}

// TestGetTransactionStatus_Success tests successful status check
func TestGetTransactionStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/status-transaction.php" {
			t.Errorf("Expected /status-transaction.php path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"message": "Transaction status retrieved",
			"data": [
				{
					"trx_id": "TRX-12345",
					"status": "berhasil",
					"payment_status": "paid"
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

	resp, err := client.GetTransactionStatus("TRX-12345")

	if err != nil {
		t.Fatalf("GetTransactionStatus() error = %v", err)
	}

	if resp.Status != "success" {
		t.Errorf("Status = %v, want success", resp.Status)
	}

	if len(resp.Data) == 0 {
		t.Fatal("Expected at least 1 transaction")
	}

	if resp.Data[0].Status != "berhasil" {
		t.Errorf("Transaction status = %v, want berhasil", resp.Data[0].Status)
	}
}

// TestGetTransactionStatus_EmptyTrxID tests validation of empty transaction ID
func TestGetTransactionStatus_EmptyTrxID(t *testing.T) {
	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})

	_, err := client.GetTransactionStatus("")

	if err != ErrMissingMerchantRef {
		t.Errorf("Expected ErrMissingMerchantRef, got %v", err)
	}
}

// TestGetTransactionHistory_WithMerchantFilter tests merchant filter parameter
func TestGetTransactionHistory_WithMerchantFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		merchant := r.FormValue("mechant")
		if merchant != "1" {
			t.Errorf("Expected mechant '1', got %s", merchant)
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

	_, err := client.GetTransactionHistory(TransactionHistoryRequest{
		MerchantFilter: 1,
	})

	if err != nil {
		t.Fatalf("GetTransactionHistory() error = %v", err)
	}
}

// TestGetTransactionHistory_RequestParams tests that all request parameters are sent correctly
func TestGetTransactionHistory_RequestParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		apiID := r.FormValue("api_id")
		method := r.FormValue("method")

		if apiID != "TEST-12345" {
			t.Errorf("Expected api_id TEST-12345, got %s", apiID)
		}
		if method != "transaction" {
			t.Errorf("Expected method 'transaction', got %s", method)
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

	_, err := client.GetTransactionHistory(TransactionHistoryRequest{})

	if err != nil {
		t.Fatalf("GetTransactionHistory() error = %v", err)
	}
}

// TestGetTransactionStatus_RequestParams tests that correct parameters are sent for status check
func TestGetTransactionStatus_RequestParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}

		apiID := r.FormValue("api_id")
		method := r.FormValue("method")
		trxID := r.FormValue("trx_id")

		if apiID != "TEST-12345" {
			t.Errorf("Expected api_id TEST-12345, got %s", apiID)
		}
		if method != "status" {
			t.Errorf("Expected method 'status', got %s", method)
		}
		if trxID != "TRX-999" {
			t.Errorf("Expected trx_id 'TRX-999', got %s", trxID)
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

	_, err := client.GetTransactionStatus("TRX-999")

	if err != nil {
		t.Fatalf("GetTransactionStatus() error = %v", err)
	}
}

// TestGetTransactionsByStatus_Pending tests getting pending transactions
func TestGetTransactionsByStatus_Pending(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "Pending transactions",
			"data": [
				{"trx_id": "TRX-PENDING-1", "status": "pending"},
				{"trx_id": "TRX-PENDING-2", "status": "pending"}
			]
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.GetTransactionsByStatus("pending")

	if err != nil {
		t.Fatalf("GetTransactionsByStatus() error = %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 pending transactions, got %d", len(resp.Data))
	}
}

// TestGetTransactionsByStatus_Expired tests getting expired transactions
func TestGetTransactionsByStatus_Expired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "Expired transactions",
			"data": [
				{"trx_id": "TRX-EXP-1", "status": "expired"}
			]
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.GetTransactionsByStatus("expired")

	if err != nil {
		t.Fatalf("GetTransactionsByStatus() error = %v", err)
	}

	if len(resp.Data) != 1 {
		t.Errorf("Expected 1 expired transaction, got %d", len(resp.Data))
	}
}

// TestGetTransactionStatus_APIErrorResponse tests error handling for status check
func TestGetTransactionStatus_APIErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{
			"status": "error",
			"message": "Transaction not found"
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	_, err := client.GetTransactionStatus("TRX-NONEXISTENT")

	if err == nil {
		t.Error("Expected error for non-existent transaction")
	}
}

// TestGetTransactionHistory_EmptyResult tests handling of empty transaction list
func TestGetTransactionHistory_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"message": "No transactions found",
			"data": []
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(Config{
		APIID:  "TEST-12345",
		APIKey: "test-key-12345",
	})
	client.baseURL = server.URL + "/"

	resp, err := client.GetTransactionHistory(TransactionHistoryRequest{})

	if err != nil {
		t.Fatalf("GetTransactionHistory() error = %v", err)
	}

	if len(resp.Data) != 0 {
		t.Errorf("Expected 0 transactions, got %d", len(resp.Data))
	}
}
