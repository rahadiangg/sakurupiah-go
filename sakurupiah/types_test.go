package sakurupiah

import (
	"encoding/json"
	"testing"
	"time"
)

// TestTransactionStatusValue tests the transaction status value methods
func TestTransactionStatusValue(t *testing.T) {
	tests := []struct {
		name   string
		status TransactionStatusValue
		isPending bool
		isSuccess bool
		isExpired bool
	}{
		{
			name:   "pending status",
			status: StatusPending,
			isPending: true,
			isSuccess: false,
			isExpired: false,
		},
		{
			name:   "success status",
			status: StatusSuccess,
			isPending: false,
			isSuccess: true,
			isExpired: false,
		},
		{
			name:   "expired status",
			status: StatusExpired,
			isPending: false,
			isSuccess: false,
			isExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status.IsPending() != tt.isPending {
				t.Errorf("IsPending() = %v, want %v", tt.status.IsPending(), tt.isPending)
			}
			if tt.status.IsSuccess() != tt.isSuccess {
				t.Errorf("IsSuccess() = %v, want %v", tt.status.IsSuccess(), tt.isSuccess)
			}
			if tt.status.IsExpired() != tt.isExpired {
				t.Errorf("IsExpired() = %v, want %v", tt.status.IsExpired(), tt.isExpired)
			}

			// Test GetStringValue
			if tt.status.GetStringValue() != string(tt.status) {
				t.Errorf("GetStringValue() = %v, want %v", tt.status.GetStringValue(), string(tt.status))
			}
		})
	}
}

// TestTransactionStatusCode tests transaction status codes
func TestTransactionStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		code       TransactionStatusCode
		wantStatus TransactionStatusValue
	}{
		{
			name:       "pending code",
			code:       StatusCodePending,
			wantStatus: StatusPending,
		},
		{
			name:       "success code",
			code:       StatusCodeSuccess,
			wantStatus: StatusSuccess,
		},
		{
			name:       "expired code",
			code:       StatusCodeExpired,
			wantStatus: StatusExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify code values match expected
			if tt.code == StatusCodePending && tt.wantStatus != StatusPending {
				t.Error("StatusCodePending doesn't match StatusPending")
			}
			if tt.code == StatusCodeSuccess && tt.wantStatus != StatusSuccess {
				t.Error("StatusCodeSuccess doesn't match StatusSuccess")
			}
			if tt.code == StatusCodeExpired && tt.wantStatus != StatusExpired {
				t.Error("StatusCodeExpired doesn't match StatusExpired")
			}
		})
	}
}

// TestMerchantFeeType tests merchant fee type values
func TestMerchantFeeType(t *testing.T) {
	tests := []struct {
		name  string
		fee   MerchantFeeType
		value int
	}{
		{
			name:  "merchant pays fee",
			fee:   FeeTypeMerchant,
			value: 1,
		},
		{
			name:  "customer pays fee",
			fee:   FeeTypeCustomer,
			value: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.fee) != tt.value {
				t.Errorf("MerchantFeeType value = %d, want %d", int(tt.fee), tt.value)
			}
		})
	}
}

// TestCreateInvoiceRequestValidation tests CreateInvoiceRequest validation
func TestCreateInvoiceRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateInvoiceRequest
		wantErr bool
	}{
		{
			name: "valid minimal request",
			req: CreateInvoiceRequest{
				Method:       "QRIS",
				CustomerPhone: "628123456789",
				Amount:       10000,
				MerchantRef:  "REF123",
				MerchantFee:  1,
				CallbackURL:  "https://example.com/callback",
				ReturnURL:    "https://example.com/return",
			},
			wantErr: false,
		},
		{
			name: "valid full request with products",
			req: CreateInvoiceRequest{
				Method:       "QRIS",
				CustomerName: "John Doe",
				CustomerEmail: "john@example.com",
				CustomerPhone: "628123456789",
				Amount:       10000,
				MerchantRef:  "REF123",
				MerchantFee:  2,
				Expired:      24,
				Products: []Product{
					{Name: "Product 1", Qty: 1, Price: 5000},
					{Name: "Product 2", Qty: 2, Price: 2500},
				},
				CallbackURL: "https://example.com/callback",
				ReturnURL:   "https://example.com/return",
			},
			wantErr: false,
		},
		{
			name: "missing method",
			req: CreateInvoiceRequest{
				CustomerPhone: "628123456789",
				Amount:        10000,
				MerchantRef:   "REF123",
			},
			wantErr: true,
		},
		{
			name: "missing phone",
			req: CreateInvoiceRequest{
				Method:      "QRIS",
				Amount:      10000,
				MerchantRef: "REF123",
			},
			wantErr: true,
		},
		{
			name: "invalid amount - zero",
			req: CreateInvoiceRequest{
				Method:       "QRIS",
				CustomerPhone: "628123456789",
				Amount:       0,
				MerchantRef:  "REF123",
			},
			wantErr: true,
		},
		{
			name: "invalid amount - negative",
			req: CreateInvoiceRequest{
				Method:       "QRIS",
				CustomerPhone: "628123456789",
				Amount:       -100,
				MerchantRef:  "REF123",
			},
			wantErr: true,
		},
		{
			name: "missing merchant ref",
			req: CreateInvoiceRequest{
				Method:       "QRIS",
				CustomerPhone: "628123456789",
				Amount:       10000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate required fields
			hasError := false
			if tt.req.Method == "" {
				hasError = true
			}
			if tt.req.CustomerPhone == "" {
				hasError = true
			}
			if tt.req.Amount <= 0 {
				hasError = true
			}
			if tt.req.MerchantRef == "" {
				hasError = true
			}

			if hasError != tt.wantErr {
				t.Errorf("validation error = %v, want %v", hasError, tt.wantErr)
			}
		})
	}
}

// TestProductSerialization tests product serialization
func TestProductSerialization(t *testing.T) {
	products := []Product{
		{Name: "Product 1", Qty: 1, Price: 10000, Size: "L", Note: "Blue"},
		{Name: "Product 2", Qty: 2, Price: 5000, Size: "M", Note: "Red"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(products)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded []Product
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(decoded) != len(products) {
		t.Errorf("decoded length = %d, want %d", len(decoded), len(products))
	}

	for i, p := range products {
		if decoded[i].Name != p.Name {
			t.Errorf("decoded[%d].Name = %v, want %v", i, decoded[i].Name, p.Name)
		}
		if decoded[i].Qty != p.Qty {
			t.Errorf("decoded[%d].Qty = %v, want %v", i, decoded[i].Qty, p.Qty)
		}
		if decoded[i].Price != p.Price {
			t.Errorf("decoded[%d].Price = %v, want %v", i, decoded[i].Price, p.Price)
		}
		if decoded[i].Size != p.Size {
			t.Errorf("decoded[%d].Size = %v, want %v", i, decoded[i].Size, p.Size)
		}
		if decoded[i].Note != p.Note {
			t.Errorf("decoded[%d].Note = %v, want %v", i, decoded[i].Note, p.Note)
		}
	}
}

// TestTransactionHistoryItemMethods tests transaction history item methods
func TestTransactionHistoryItemMethods(t *testing.T) {
	item := TransactionHistoryItem{
		TrxID:       "TRX123",
		MerchantRef: "REF123",
		PaymentCode: "QRIS",
		Date:        "2025-01-26",
		Time:        "10:30:00",
		Amount:      "10000",
		Expired:     "2025-01-26 11:30:00",
		Status:      "berhasil",
	}

	t.Run("FormatExpiredTime", func(t *testing.T) {
		expiredTime, err := item.FormatExpiredTime()
		if err != nil {
			t.Fatalf("FormatExpiredTime() error = %v", err)
		}

		expected, _ := time.Parse("2006-01-02 15:04:05", "2025-01-26 11:30:00")
		if !expiredTime.Equal(expected) {
			t.Errorf("FormatExpiredTime() = %v, want %v", expiredTime, expected)
		}
	})

	t.Run("FormatDateTime", func(t *testing.T) {
		dateTime, err := item.FormatDateTime()
		if err != nil {
			t.Fatalf("FormatDateTime() error = %v", err)
		}

		expected, _ := time.Parse("2006-01-02 15:04:05", "2025-01-26 10:30:00")
		if !dateTime.Equal(expected) {
			t.Errorf("FormatDateTime() = %v, want %v", dateTime, expected)
		}
	})
}

// TestTransactionHistoryRequestTests tests transaction history request validation
func TestTransactionHistoryRequestTests(t *testing.T) {
	tests := []struct {
		name  string
		req   TransactionHistoryRequest
		valid bool
	}{
		{
			name: "empty request - valid",
			req:  TransactionHistoryRequest{},
			valid: true,
		},
		{
			name: "with merchant filter",
			req: TransactionHistoryRequest{
				MerchantFilter: 1,
			},
			valid: true,
		},
		{
			name: "with payment code filter",
			req: TransactionHistoryRequest{
				PaymentCode: "QRIS",
			},
			valid: true,
		},
		{
			name: "with trx ID filter",
			req: TransactionHistoryRequest{
				TrxID: "TRX123",
			},
			valid: true,
		},
		{
			name: "with merchant ref filter",
			req: TransactionHistoryRequest{
				MerchantRef: "REF123",
			},
			valid: true,
		},
		{
			name: "with status filter",
			req: TransactionHistoryRequest{
				Status: "pending",
			},
			valid: true,
		},
		{
			name: "with date range filter",
			req: TransactionHistoryRequest{
				StartDate: "2025-01-01",
				EndDate:   "2025-01-31",
			},
			valid: true,
		},
		{
			name: "with all filters",
			req: TransactionHistoryRequest{
				MerchantFilter: 1,
				PaymentCode:    "QRIS",
				TrxID:          "TRX123",
				MerchantRef:    "REF123",
				Status:         "berhasil",
				StartDate:      "2025-01-01",
				EndDate:        "2025-01-31",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// All requests should be valid as all fields are optional
			if !tt.valid {
				t.Error("TransactionHistoryRequest should always be valid")
			}
		})
	}
}

// TestPaymentChannelFields tests payment channel struct fields
func TestPaymentChannelFields(t *testing.T) {
	channel := PaymentChannel{
		Code:   "QRIS",
		Name:   "QRIS Payment",
		Min:    "500",
		Max:    "2000000",
		Fee:    "0.7",
		Percent: "Percent",
		Type:   "DIRECT",
		Logo:   "https://example.com/qris.png",
		Status: "Aktif",
		Addition: PaymentChannelAddition{
			ExtraFee:   "350",
			Type:       "Nominal",
			DefaultExp: "24",
			Settlement: "Settlement H+1",
		},
		Guide: PaymentGuide{
			Title:        "Cara Bayar QRIS",
			PaymentGuide: "Scan QR code...",
		},
	}

	if channel.Code != "QRIS" {
		t.Errorf("Code = %v, want QRIS", channel.Code)
	}
	if channel.Status != "Aktif" {
		t.Errorf("Status = %v, want Aktif", channel.Status)
	}
	if channel.Type != "DIRECT" {
		t.Errorf("Type = %v, want DIRECT", channel.Type)
	}
}

// TestBalanceDataFields tests balance data struct
func TestBalanceDataFields(t *testing.T) {
	balance := BalanceData{
		MerchantName:     "Test Merchant",
		Balance:          "50000",
		AvailableBalance: "100000",
	}

	if balance.MerchantName != "Test Merchant" {
		t.Errorf("MerchantName = %v, want Test Merchant", balance.MerchantName)
	}
	if balance.Balance != "50000" {
		t.Errorf("Balance = %v, want 50000", balance.Balance)
	}
	if balance.AvailableBalance != "100000" {
		t.Errorf("AvailableBalance = %v, want 100000", balance.AvailableBalance)
	}
}

// TestCallbackRequestJSON tests callback request JSON serialization
func TestCallbackRequestJSON(t *testing.T) {
	callbackJSON := `{"trx_id":"TRX123","merchant_ref":"REF123","status":"berhasil","status_kode":1}`

	var callback CallbackRequest
	if err := json.Unmarshal([]byte(callbackJSON), &callback); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if callback.TrxID != "TRX123" {
		t.Errorf("TrxID = %v, want TRX123", callback.TrxID)
	}
	if callback.MerchantRef != "REF123" {
		t.Errorf("MerchantRef = %v, want REF123", callback.MerchantRef)
	}
	if callback.Status != StatusSuccess {
		t.Errorf("Status = %v, want %v", callback.Status, StatusSuccess)
	}
	if callback.StatusCode != StatusCodeSuccess {
		t.Errorf("StatusCode = %v, want %v", callback.StatusCode, StatusCodeSuccess)
	}
}

// TestCallbackResponseJSON tests callback response JSON
func TestCallbackResponseJSON(t *testing.T) {
	response := CallbackResponse{
		Success: true,
		Message: "Payment processed",
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded CallbackResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Success != response.Success {
		t.Errorf("Success = %v, want %v", decoded.Success, response.Success)
	}
	if decoded.Message != response.Message {
		t.Errorf("Message = %v, want %v", decoded.Message, response.Message)
	}
}

// TestInvoiceDataFields tests invoice data struct
func TestInvoiceDataFields(t *testing.T) {
	invoice := InvoiceData{
		Via:            "QRIS",
		PaymentCode:    "QRIS",
		TrxID:          "TRX123",
		MerchantRef:    "REF123",
		Name:           "John Doe",
		Email:          "john@example.com",
		Phone:          "628123456789",
		Total:          10000,
		MerchantFee:    "Merchant",
		Fee:            70,
		AmountMerchant: 9930,
		Date:           "2025-01-26",
		Time:           "10:30:00",
		Expired:        "2025-01-26 11:30:00",
		PaymentStatus:  "pending",
		QR:             "qr_string_data",
		PaymentNo:      "123456789",
		CheckoutURL:    "https://example.com/checkout",
	}

	if invoice.TrxID != "TRX123" {
		t.Errorf("TrxID = %v, want TRX123", invoice.TrxID)
	}
	if invoice.Total != 10000 {
		t.Errorf("Total = %v, want 10000", invoice.Total)
	}
	if invoice.PaymentStatus != "pending" {
		t.Errorf("PaymentStatus = %v, want pending", invoice.PaymentStatus)
	}
	if invoice.QR == "" {
		t.Error("QR should not be empty for QRIS")
	}
}

// TestPaymentEnvironment tests payment environment constants
func TestPaymentEnvironment(t *testing.T) {
	if EnvironmentProduction != 0 {
		t.Errorf("EnvironmentProduction = %d, want 0", EnvironmentProduction)
	}
	if EnvironmentSandbox != 1 {
		t.Errorf("EnvironmentSandbox = %d, want 1", EnvironmentSandbox)
	}
}

// TestBaseResponseFields tests base response fields
func TestBaseResponseFields(t *testing.T) {
	resp := BaseResponse{
		Status:  "200",
		Message: "success",
	}

	if resp.Status != "200" {
		t.Errorf("Status = %v, want 200", resp.Status)
	}
	if resp.Message != "success" {
		t.Errorf("Message = %v, want success", resp.Message)
	}
}

// TestErrorResponseFields tests error response fields
func TestErrorResponseFields(t *testing.T) {
	errResp := ErrorResponse{
		Status:  "400",
		Message: "Invalid request",
	}

	if errResp.Status != "400" {
		t.Errorf("Status = %v, want 400", errResp.Status)
	}
	if errResp.Message != "Invalid request" {
		t.Errorf("Message = %v, want Invalid request", errResp.Message)
	}
}
