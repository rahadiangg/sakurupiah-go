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

// TestFlexibleStringUnmarshalJSON tests FlexibleString JSON unmarshaling
func TestFlexibleStringUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    string
		wantErr bool
	}{
		{
			name:    "string value",
			json:    `"hello"`,
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "numeric value",
			json:    `12345`,
			want:    "12345",
			wantErr: false,
		},
		{
			name:    "numeric value as string",
			json:    `"67890"`,
			want:    "67890",
			wantErr: false,
		},
		{
			name:    "zero number",
			json:    `0`,
			want:    "0",
			wantErr: false,
		},
		{
			name:    "negative number",
			json:    `-123`,
			want:    "-123",
			wantErr: false,
		},
		{
			name:    "floating point number",
			json:    `12.34`,
			want:    "12.34",
			wantErr: false,
		},
		{
			name:    "boolean",
			json:    `true`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "null",
			json:    `null`,
			want:    "",
			wantErr: false, // null becomes empty string
		},
		{
			name:    "array",
			json:    `[]`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "object",
			json:    `{}`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fs FlexibleString
			err := json.Unmarshal([]byte(tt.json), &fs)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UnmarshalJSON() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalJSON() unexpected error = %v", err)
				return
			}

			if fs.String() != tt.want {
				t.Errorf("String() = %v, want %v", fs.String(), tt.want)
			}
		})
	}
}

// TestFlexibleStringString tests String() method
func TestFlexibleStringString(t *testing.T) {
	tests := []struct {
		name string
		fs   FlexibleString
		want string
	}{
		{
			name: "simple string",
			fs:   FlexibleString("test"),
			want: "test",
		},
		{
			name: "numeric string",
			fs:   FlexibleString("12345"),
			want: "12345",
		},
		{
			name: "empty string",
			fs:   FlexibleString(""),
			want: "",
		},
		{
			name: "string with special chars",
			fs:   FlexibleString("hello-world_123"),
			want: "hello-world_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fs.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFlexibleStringInt64 tests Int64() method
func TestFlexibleStringInt64(t *testing.T) {
	tests := []struct {
		name    string
		fs      FlexibleString
		want    int64
		wantErr bool
	}{
		{
			name:    "valid positive number",
			fs:      FlexibleString("12345"),
			want:    12345,
			wantErr: false,
		},
		{
			name:    "valid negative number",
			fs:      FlexibleString("-12345"),
			want:    -12345,
			wantErr: false,
		},
		{
			name:    "zero",
			fs:      FlexibleString("0"),
			want:    0,
			wantErr: false,
		},
		{
			name:    "large number",
			fs:      FlexibleString("9223372036854775807"),
			want:    9223372036854775807,
			wantErr: false,
		},
		{
			name:    "invalid string",
			fs:      FlexibleString("abc"),
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty string",
			fs:      FlexibleString(""),
			want:    0,
			wantErr: true,
		},
		{
			name:    "string with spaces",
			fs:      FlexibleString("123 456"),
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fs.Int64()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Int64() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Int64() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFlexibleStringMustInt64 tests MustInt64() method
func TestFlexibleStringMustInt64(t *testing.T) {
	tests := []struct {
		name     string
		fs       FlexibleString
		want     int64
		wantPanic bool
	}{
		{
			name:     "valid positive number",
			fs:       FlexibleString("12345"),
			want:     12345,
			wantPanic: false,
		},
		{
			name:     "valid negative number",
			fs:       FlexibleString("-12345"),
			want:     -12345,
			wantPanic: false,
		},
		{
			name:     "zero",
			fs:       FlexibleString("0"),
			want:     0,
			wantPanic: false,
		},
		{
			name:     "invalid string",
			fs:       FlexibleString("abc"),
			want:     0,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("MustInt64() should have panicked")
					}
				}()
			}

			got := tt.fs.MustInt64()
			if !tt.wantPanic && got != tt.want {
				t.Errorf("MustInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFlexibleInt64UnmarshalJSON tests FlexibleInt64 JSON unmarshaling
func TestFlexibleInt64UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    int64
		wantErr bool
	}{
		{
			name:    "numeric value",
			json:    `12345`,
			want:    12345,
			wantErr: false,
		},
		{
			name:    "numeric string",
			json:    `"67890"`,
			want:    67890,
			wantErr: false,
		},
		{
			name:    "negative numeric",
			json:    `-123`,
			want:    -123,
			wantErr: false,
		},
		{
			name:    "negative numeric string",
			json:    `"-456"`,
			want:    -456,
			wantErr: false,
		},
		{
			name:    "zero",
			json:    `0`,
			want:    0,
			wantErr: false,
		},
		{
			name:    "large number",
			json:    `9223372036854775807`,
			want:    9223372036854775807,
			wantErr: false,
		},
		{
			name:    "invalid string",
			json:    `"abc"`,
			want:    0,
			wantErr: true,
		},
		{
			name:    "boolean",
			json:    `true`,
			want:    0,
			wantErr: true,
		},
		{
			name:    "null",
			json:    `null`,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fi FlexibleInt64
			err := json.Unmarshal([]byte(tt.json), &fi)

			if tt.wantErr {
				if err == nil {
					t.Errorf("UnmarshalJSON() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("UnmarshalJSON() unexpected error = %v", err)
				return
			}

			if fi.Int64() != tt.want {
				t.Errorf("Int64() = %v, want %v", fi.Int64(), tt.want)
			}
		})
	}
}

// TestFlexibleInt64Int64 tests Int64() method
func TestFlexibleInt64Int64(t *testing.T) {
	tests := []struct {
		name string
		fi   FlexibleInt64
		want int64
	}{
		{
			name: "positive number",
			fi:   FlexibleInt64(12345),
			want: 12345,
		},
		{
			name: "negative number",
			fi:   FlexibleInt64(-12345),
			want: -12345,
		},
		{
			name: "zero",
			fi:   FlexibleInt64(0),
			want: 0,
		},
		{
			name: "max int64",
			fi:   FlexibleInt64(9223372036854775807),
			want: 9223372036854775807,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fi.Int64(); got != tt.want {
				t.Errorf("Int64() = %v, want %v", got, tt.want)
			}
		})
	}
}
