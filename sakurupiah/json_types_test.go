package sakurupiah

import (
	"encoding/json"
	"testing"
)

// TestFlexibleStatusCodeUnmarshal tests that FlexibleStatusCode can unmarshal
// both string and numeric JSON values.
//
// This is critical because the Sakurupiah API sends status_kode as a string
// (e.g., "1", "0", "-2") despite documenting it as an integer.
func TestFlexibleStatusCodeUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantCode TransactionStatusCode
		wantInt  int
	}{
		{
			name:     "numeric success (1)",
			json:     `1`,
			wantCode: StatusCodeSuccess,
			wantInt:  1,
		},
		{
			name:     "string success (\"1\")",
			json:     `"1"`,
			wantCode: StatusCodeSuccess,
			wantInt:  1,
		},
		{
			name:     "numeric pending (0)",
			json:     `0`,
			wantCode: StatusCodePending,
			wantInt:  0,
		},
		{
			name:     "string pending (\"0\")",
			json:     `"0"`,
			wantCode: StatusCodePending,
			wantInt:  0,
		},
		{
			name:     "numeric expired (-2)",
			json:     `-2`,
			wantCode: StatusCodeExpired,
			wantInt:  -2,
		},
		{
			name:     "string expired (\"-2\")",
			json:     `"-2"`,
			wantCode: StatusCodeExpired,
			wantInt:  -2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fs FlexibleStatusCode
			if err := json.Unmarshal([]byte(tt.json), &fs); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if fs.Int() != tt.wantInt {
				t.Errorf("Int() = %v, want %v", fs.Int(), tt.wantInt)
			}

			if fs.TransactionStatusCode() != tt.wantCode {
				t.Errorf("TransactionStatusCode() = %v, want %v", fs.TransactionStatusCode(), tt.wantCode)
			}
		})
	}
}

// TestFlexibleStatusCodeUnmarshalInvalid tests invalid inputs.
func TestFlexibleStatusCodeUnmarshalInvalid(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "invalid string",
			json:    `"abc"`,
			wantErr: true,
		},
		{
			name:    "boolean",
			json:    `true`,
			wantErr: true,
		},
		{
			name:    "null",
			json:    `null`,
			wantErr: true,
		},
		{
			name:    "array",
			json:    `[]`,
			wantErr: true,
		},
		{
			name:    "object",
			json:    `{}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fs FlexibleStatusCode
			err := json.Unmarshal([]byte(tt.json), &fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFlexibleStatusCodeMarshal tests that FlexibleStatusCode marshals as an integer.
func TestFlexibleStatusCodeMarshal(t *testing.T) {
	tests := []struct {
		name string
		code FlexibleStatusCode
		want string
	}{
		{
			name: "success",
			code: FlexibleStatusCode(StatusCodeSuccess),
			want: `1`,
		},
		{
			name: "pending",
			code: FlexibleStatusCode(StatusCodePending),
			want: `0`,
		},
		{
			name: "expired",
			code: FlexibleStatusCode(StatusCodeExpired),
			want: `-2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.code)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			got := string(data)
			if got != tt.want {
				t.Errorf("Marshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFlexibleStatusCodeRoundTrip tests marshal/unmarshal round trip.
func TestFlexibleStatusCodeRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		code FlexibleStatusCode
	}{
		{
			name: "success",
			code: FlexibleStatusCode(StatusCodeSuccess),
		},
		{
			name: "pending",
			code: FlexibleStatusCode(StatusCodePending),
		},
		{
			name: "expired",
			code: FlexibleStatusCode(StatusCodeExpired),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.code)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			var decoded FlexibleStatusCode
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if decoded != tt.code {
				t.Errorf("RoundTrip = %v, want %v", decoded, tt.code)
			}
		})
	}
}

// TestCallbackRequestStringStatusCode tests that CallbackRequest can handle
// string status_kode values from the Sakurupiah API.
//
// This test directly addresses the bug where the API sends status_kode as
// a string despite documenting it as an integer.
func TestCallbackRequestStringStatusCode(t *testing.T) {
	tests := []struct {
		name           string
		jsonPayload    string
		wantStatusCode TransactionStatusCode
		wantInt        int
	}{
		{
			name:           "string status_kode success (\"1\")",
			jsonPayload:    `{"trx_id":"TRX123","merchant_ref":"REF123","status":"berhasil","status_kode":"1"}`,
			wantStatusCode: StatusCodeSuccess,
			wantInt:        1,
		},
		{
			name:           "string status_kode pending (\"0\")",
			jsonPayload:    `{"trx_id":"TRX123","merchant_ref":"REF123","status":"pending","status_kode":"0"}`,
			wantStatusCode: StatusCodePending,
			wantInt:        0,
		},
		{
			name:           "string status_kode expired (\"-2\")",
			jsonPayload:    `{"trx_id":"TRX123","merchant_ref":"REF123","status":"expired","status_kode":"-2"}`,
			wantStatusCode: StatusCodeExpired,
			wantInt:        -2,
		},
		{
			name:           "numeric status_kode success (1)",
			jsonPayload:    `{"trx_id":"TRX123","merchant_ref":"REF123","status":"berhasil","status_kode":1}`,
			wantStatusCode: StatusCodeSuccess,
			wantInt:        1,
		},
		{
			name:           "numeric status_kode pending (0)",
			jsonPayload:    `{"trx_id":"TRX123","merchant_ref":"REF123","status":"pending","status_kode":0}`,
			wantStatusCode: StatusCodePending,
			wantInt:        0,
		},
		{
			name:           "numeric status_kode expired (-2)",
			jsonPayload:    `{"trx_id":"TRX123","merchant_ref":"REF123","status":"expired","status_kode":-2}`,
			wantStatusCode: StatusCodeExpired,
			wantInt:        -2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callback CallbackRequest
			if err := json.Unmarshal([]byte(tt.jsonPayload), &callback); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if callback.StatusCode.TransactionStatusCode() != tt.wantStatusCode {
				t.Errorf("StatusCode = %v, want %v", callback.StatusCode.TransactionStatusCode(), tt.wantStatusCode)
			}

			if callback.StatusCode.Int() != tt.wantInt {
				t.Errorf("StatusCode.Int() = %v, want %v", callback.StatusCode.Int(), tt.wantInt)
			}
		})
	}
}

// TestCallbackRequestMarshalCallbackRequest tests that CallbackRequest
// marshals correctly for outgoing data.
func TestCallbackRequestMarshalCallbackRequest(t *testing.T) {
	callback := CallbackRequest{
		TrxID:       "TRX123",
		MerchantRef: "REF123",
		Status:      StatusSuccess,
		StatusCode:  FlexibleStatusCode(StatusCodeSuccess),
	}

	data, err := json.Marshal(callback)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Check that status_kode is marshaled as a number, not string
	statusKode, ok := decoded["status_kode"].(float64)
	if !ok {
		t.Fatalf("status_kode should be a number in JSON, got %T", decoded["status_kode"])
	}

	if int(statusKode) != int(StatusCodeSuccess) {
		t.Errorf("status_kode = %v, want %v", int(statusKode), StatusCodeSuccess)
	}
}
