package sakurupiah

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexibleString is a string type that can unmarshal from both string and JSON numbers
type FlexibleString string

// UnmarshalJSON implements custom JSON unmarshaling for FlexibleString
func (fs *FlexibleString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fs = FlexibleString(s)
		return nil
	}

	// Try to unmarshal as number
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*fs = FlexibleString(n.String())
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into FlexibleString", string(data))
}

// String returns the string value
func (fs FlexibleString) String() string {
	return string(fs)
}

// Int64 converts FlexibleString to int64
func (fs FlexibleString) Int64() (int64, error) {
	return strconv.ParseInt(string(fs), 10, 64)
}

// MustInt64 converts FlexibleString to int64, panics on error
func (fs FlexibleString) MustInt64() int64 {
	val, err := fs.Int64()
	if err != nil {
		panic(err)
	}
	return val
}

// FlexibleInt64 is an int64 type that can unmarshal from both string and JSON numbers
type FlexibleInt64 int64

// UnmarshalJSON implements custom JSON unmarshaling for FlexibleInt64
func (fi *FlexibleInt64) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as number first
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		val, err := n.Int64()
		if err != nil {
			return err
		}
		*fi = FlexibleInt64(val)
		return nil
	}

	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*fi = FlexibleInt64(val)
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into FlexibleInt64", string(data))
}

// Int64 returns the int64 value
func (fi FlexibleInt64) Int64() int64 {
	return int64(fi)
}

// FlexibleStatusCode is a TransactionStatusCode type that can unmarshal from both string and JSON numbers.
// This is needed because the Sakurupiah API sends status_kode as a string (e.g., "1") despite
// documenting it as an integer.
type FlexibleStatusCode TransactionStatusCode

// UnmarshalJSON implements custom JSON unmarshaling for FlexibleStatusCode.
// It accepts both string ("1", "0", "-2") and numeric (1, 0, -2) formats.
func (fs *FlexibleStatusCode) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as number first
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		val, err := n.Int64()
		if err != nil {
			return err
		}
		*fs = FlexibleStatusCode(val)
		return nil
	}

	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse status code from string %q: %w", s, err)
		}
		*fs = FlexibleStatusCode(val)
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into FlexibleStatusCode", string(data))
}

// Int returns the int value of the status code.
func (fs FlexibleStatusCode) Int() int {
	return int(fs)
}

// TransactionStatusCode returns the TransactionStatusCode value.
func (fs FlexibleStatusCode) TransactionStatusCode() TransactionStatusCode {
	return TransactionStatusCode(fs)
}

// MarshalJSON implements custom JSON marshaling for FlexibleStatusCode.
// It marshals as an integer for outgoing requests.
func (fs FlexibleStatusCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(fs))
}
