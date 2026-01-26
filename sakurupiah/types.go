package sakurupiah

import "time"

// ============================================================
// BASE TYPES
// ============================================================

// BaseResponse represents the standard API response structure.
//
// All Sakurupiah API responses include a status field and message field
// to indicate the result of the API operation.
//
//   - Status: Typically "success" or "error"
//   - Message: Human-readable description of the result
type BaseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ============================================================
// CREATE INVOICE TYPES
// ============================================================

// CreateInvoiceRequest represents the request to create a new invoice.
//
// This struct contains all the parameters needed to create a payment invoice
// with the Sakurupiah payment gateway. All required fields must be populated
// before calling CreateInvoice.
//
// # Required Fields
//   - Method: Payment channel code (e.g., "QRIS", "BCAVA")
//   - CustomerPhone: Customer phone number
//   - Amount: Transaction amount in IDR
//   - MerchantRef: Your unique transaction reference
//   - CallbackURL: URL for payment status notifications
//   - ReturnURL: URL to redirect customer after payment
//
// # Optional Fields
//   - CustomerName: Customer name (optional)
//   - CustomerEmail: Customer email (optional)
//   - MerchantFee: Fee payer (1=merchant, 2=customer, default=1)
//   - Expired: Payment expiration in hours (default depends on channel)
//   - Products: Array of product items (optional)
type CreateInvoiceRequest struct {
	// Method is the payment channel code (e.g., "QRIS", "BCAVA", "DANA")
	Method string `json:"method"`
	// Customer details
	CustomerName  string `json:"name,omitempty"`
	CustomerEmail string `json:"email,omitempty"`
	CustomerPhone string `json:"phone"`
	// Amount is the total amount in IDR
	Amount int64 `json:"amount"`
	// MerchantFee: 1 = merchant pays, 2 = customer pays, default = merchant pays
	MerchantFee int `json:"merchant_fee"`
	// MerchantRef is a unique reference from merchant system
	MerchantRef string `json:"merchant_ref"`
	// Expired is payment expiration in hours (optional, default depends on channel)
	Expired int `json:"expired,omitempty"`
	// Products details (optional)
	Products []Product `json:"products,omitempty"`
	// CallbackURL is where payment status updates will be sent
	CallbackURL string `json:"callback_url"`
	// ReturnURL is where customer will be redirected after payment
	ReturnURL string `json:"return_url"`
	// Signature is the HMAC-SHA256 signature
	Signature string `json:"-"`
}

// Product represents a product item in the invoice.
//
// When creating an invoice with product details, each product should include
// the name, quantity, and price. Size and note are optional.
//
// # Example
//
//	products := []sakurupiah.Product{
//	    {Name: "T-Shirt", Qty: 1, Price: 50000, Size: "L", Note: "Blue"},
//	    {Name: "Pants", Qty: 2, Price: 75000, Size: "M", Note: "Black"},
//	}
type Product struct {
	Name  string `json:"name"`
	Qty   int    `json:"qty"`
	Price int64  `json:"price"`
	Size  string `json:"size,omitempty"`
	Note  string `json:"note,omitempty"`
}

// InvoiceData represents the invoice data in the response.
//
// This struct contains all the details of a created invoice including
// payment information, customer details, fees, and payment instructions.
//
// # Key Fields
//   - TrxID: Unique transaction ID from Sakurupiah
//   - MerchantRef: Your reference for the transaction
//   - PaymentStatus: Current payment status ("pending", "berhasil", "expired")
//   - CheckoutURL: URL to redirect customer for payment (for REDIRECT type channels)
//   - QR: QR code string for QRIS payments
//   - PaymentNo: Payment number/account for DIRECT payment methods
type InvoiceData struct {
	Via           string `json:"via"`
	PaymentCode   string `json:"payment_kode"`
	TrxID         string `json:"trx_id"`
	MerchantRef   string `json:"merchant_ref"`
	Name          string `json:"nama"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Total         int64  `json:"total"`
	MerchantFee   string `json:"merchant_fee"`
	Fee           int64  `json:"fee"`
	AmountMerchant int64  `json:"amount_merchant"`
	Date          string `json:"date"`
	Time          string `json:"time"`
	Expired       string `json:"expired"`
	PaymentStatus string          `json:"payment_status"`
	QR            string          `json:"qr,omitempty"`           // For QRIS payments
	PaymentNo     FlexibleString `json:"payment_no,omitempty"`   // Payment number (can be string or number in response)
	CheckoutURL   string          `json:"checkout_url,omitempty"` // Checkout URL
}

// ProductResponse represents a product in the response.
//
// Similar to Product but uses FlexibleString for quantity since the API
// may return either a string or number for this field.
type ProductResponse struct {
	Name  string          `json:"nama_produk"`
	Qty   FlexibleString `json:"qty"`  // Can be string or number in response
	Price int64           `json:"harga"`
	Size  string          `json:"size,omitempty"`
	Note  string          `json:"note,omitempty"`
}

// CreateInvoiceResponse represents the response from creating an invoice.
//
// Contains the base response (status/message) along with the invoice data
// and optionally the product details that were included in the invoice.
type CreateInvoiceResponse struct {
	BaseResponse
	Data    []InvoiceData     `json:"data"`
	Product []ProductResponse `json:"produk,omitempty"`
}

// ============================================================
// LIST PAYMENT CHANNELS TYPES
// ============================================================

// ListPaymentChannelsResponse represents the response from listing payment channels.
//
// Contains the base response along with an array of available payment channels.
type ListPaymentChannelsResponse struct {
	BaseResponse
	Data []PaymentChannel `json:"data"`
}

// PaymentChannel represents a payment channel.
//
// Contains all information about a payment method including availability,
// transaction limits, fees, and payment instructions.
//
// # Key Fields
//   - Code: Payment method code for API requests (e.g., "QRIS", "BCAVA")
//   - Name: Display name of the payment method
//   - Status: "Aktif" (active) or "Offline" (inactive)
//   - Min/Max: Transaction amount limits
//   - Fee: Transaction fee structure
//   - Type: "DIRECT" (customer pays directly) or "REDIRECT" (redirect to payment page)
//   - Guide: Payment instructions for customers
type PaymentChannel struct {
	Code     string              `json:"kode"`
	Name     string              `json:"nama"`
	Min      string              `json:"minimal"`
	Max      string              `json:"maksimal"`
	Fee      string              `json:"biaya"`
	Percent  string              `json:"percent"`
	Type     string              `json:"tipe"` // DIRECT or REDIRECT
	Logo     string              `json:"logo"`
	Status   string              `json:"status"` // Aktif or Offline
	Addition PaymentChannelAddition `json:"addition"`
	Guide    PaymentGuide        `json:"guide"`
}

// PaymentChannelAddition contains additional payment channel info.
//
// Provides extended information about a payment channel including
// fee structure, default expiration time, and settlement schedule.
type PaymentChannelAddition struct {
	ExtraFee     string `json:"tambahan_biaya"`
	Type         string `json:"jenis"`      // Nominal or Percent
	DefaultExp   string `json:"default_expired"`
	Settlement   string `json:"settlement"` // Settlement time
}

// PaymentGuide contains payment instructions.
//
// Provides human-readable payment instructions that can be displayed
// to customers to guide them through the payment process.
type PaymentGuide struct {
	Title          string `json:"title"`
	PaymentGuide   string `json:"payment_guide"`
}

// ============================================================
// CHECK BALANCE TYPES
// ============================================================

// CheckBalanceResponse represents the response from checking balance.
//
// Contains the base response along with merchant balance information.
type CheckBalanceResponse struct {
	BaseResponse
	Data BalanceData `json:"data"`
}

// BalanceData represents balance information.
//
// Contains the merchant's balance details including pending settlement
// balance and available balance that can be withdrawn.
//
// # Balance Types
//   - Balance: Pending settlement balance (transactions awaiting settlement)
//   - AvailableBalance: Ready-to-use balance that has been settled
//
// Settlement times vary by payment method (typically H+1 to H+3).
type BalanceData struct {
	MerchantName    string `json:"nama_merchant"`
	Balance         string `json:"balance"`          // Pending settlement balance
	AvailableBalance string `json:"saldo_tersedia"`  // Available balance
}

// ============================================================
// TRANSACTION HISTORY TYPES
// ============================================================

// TransactionHistoryRequest represents the request for transaction history.
//
// Use this to filter and query transaction history. All fields are optional -
// leave them empty to retrieve all transactions without filters.
//
// # Filter Options
//   - MerchantFilter: Filter by merchant (1=current only, 0/empty=all)
//   - PaymentCode: Filter by payment channel code
//   - TrxID: Filter by Sakurupiah transaction ID
//   - MerchantRef: Filter by your merchant reference
//   - Status: Filter by status ("pending", "berhasil", "expired")
//   - StartDate/EndDate: Filter by date range (format: "YYYY-MM-DD")
type TransactionHistoryRequest struct {
	// Filter by merchant (1 = current merchant only, 0 or empty = all merchants)
	MerchantFilter int `json:"mechant,omitempty"`
	// Filter by payment channel code
	PaymentCode string `json:"payment_kode,omitempty"`
	// Filter by transaction ID from Sakurupiah
	TrxID string `json:"trx_id,omitempty"`
	// Filter by merchant reference
	MerchantRef string `json:"merchant_ref,omitempty"`
	// Filter by status: pending, berhasil, expired
	Status string `json:"status,omitempty"`
	// Filter by date range (format: YYYY-MM-DD)
	StartDate string `json:"tanggal_awal,omitempty"`
	EndDate   string `json:"tanggal_akhir,omitempty"`
}

// TransactionHistoryResponse represents the response from transaction history.
//
// Contains the base response along with an array of transaction items.
type TransactionHistoryResponse struct {
	BaseResponse
	Data []TransactionHistoryItem `json:"data"`
}

// TransactionHistoryItem represents a transaction in history.
//
// Contains summary information about a transaction including its status,
// amount, dates, and payment method.
//
// # Helper Methods
//   - FormatExpiredTime(): Parse the expiration time into time.Time
//   - FormatDateTime(): Parse the transaction datetime into time.Time
type TransactionHistoryItem struct {
	TrxID        string `json:"trx_id"`
	MerchantRef  string `json:"merchant_ref"`
	PaymentCode  string `json:"payment_kode"`
	Date         string `json:"tanggal"`
	Time         string `json:"waktu"`
	Amount       string `json:"amount"`
	Expired      string `json:"expired"`
	Status       string `json:"status"`
}

// ============================================================
// TRANSACTION STATUS TYPES
// ============================================================

// TransactionStatusResponse represents the response from checking transaction status.
//
// Contains the base response along with an array of transaction status information.
type TransactionStatusResponse struct {
	BaseResponse
	Data []TransactionStatus `json:"data"`
}

// TransactionStatus represents transaction status.
//
// Simple struct containing only the status field which can be
// "pending", "berhasil" (success), or "expired".
type TransactionStatus struct {
	Status string `json:"status"` // pending, berhasil, expired
}

// ============================================================
// CALLBACK TYPES
// ============================================================

// TransactionStatusValue represents the status of a transaction.
//
// Use the constants StatusPending, StatusSuccess, and StatusExpired
// for comparing transaction statuses.
type TransactionStatusValue string

const (
	// StatusPending represents pending transaction
	StatusPending TransactionStatusValue = "pending"
	// StatusSuccess represents successful transaction
	StatusSuccess TransactionStatusValue = "berhasil"
	// StatusExpired represents expired transaction
	StatusExpired TransactionStatusValue = "expired"
)

// TransactionStatusCode represents the status code.
//
// Numeric representation of transaction status where:
//   - 0 = pending
//   - 1 = success
//   - -2 = expired
type TransactionStatusCode int

const (
	// StatusCodePending is the code for pending status
	StatusCodePending TransactionStatusCode = 0
	// StatusCodeSuccess is the code for successful status
	StatusCodeSuccess TransactionStatusCode = 1
	// StatusCodeExpired is the code for expired status
	StatusCodeExpired TransactionStatusCode = -2
)

// CallbackRequest represents the callback data from Sakurupiah.
//
// This struct contains the data sent by Sakurupiah when a payment
// status changes. Always verify the signature before processing callbacks.
//
// # Security
//
// The RawPayload field contains the original JSON bytes used for
// signature verification. This is automatically populated when using
// VerifyAndParseCallback().
type CallbackRequest struct {
	TrxID        string                 `json:"trx_id"`
	MerchantRef  string                 `json:"merchant_ref"`
	Status       TransactionStatusValue `json:"status"`
	StatusCode   TransactionStatusCode  `json:"status_kode"`
	RawPayload   []byte                 `json:"-"` // Raw JSON payload for signature verification
}

// CallbackHeaders represents the headers from callback.
//
// Contains the signature and event type from callback HTTP headers.
type CallbackHeaders struct {
	Signature string `json:"X-Callback-Signature"`
	Event     string `json:"X-Callback-Event"`
}

// CallbackResponse represents the response to callback.
//
// Send this response back to Sakurupiah to acknowledge receipt of
// the payment callback. Use success=true to indicate successful processing.
type CallbackResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ============================================================
// HELPER TYPES
// ============================================================

// MerchantFeeType represents who pays the transaction fee.
//
// Use FeeTypeMerchant or FeeTypeCustomer when specifying merchant_fee
// in CreateInvoiceRequest.
type MerchantFeeType int

const (
	// FeeTypeMerchant means merchant pays the fee
	FeeTypeMerchant MerchantFeeType = 1
	// FeeTypeCustomer means customer pays the fee
	FeeTypeCustomer MerchantFeeType = 2
)

// PaymentEnvironment represents the API environment.
//
// EnvironmentProduction and EnvironmentSandbox can be used to
// represent the current API environment in configuration.
type PaymentEnvironment int

const (
	// EnvironmentProduction uses production API
	EnvironmentProduction PaymentEnvironment = 0
	// EnvironmentSandbox uses sandbox API
	EnvironmentSandbox PaymentEnvironment = 1
)

// ============================================================
// PAYMENT METHOD CONSTANTS
// ============================================================

const (
	// QRIS Payment Methods
	MethodQRIS  = "QRIS"
	MethodQRIS2 = "QRIS2"
	MethodQRISM = "QRISM"
	MethodQRISC = "QRISC"

	// Virtual Account Payment Methods
	MethodBCAVA     = "BCAVA"
	MethodBRIVA     = "BRIVA"
	MethodBNIVA     = "BNIVA"
	MethodBAGVA     = "BAGVA"
	MethodBNCVA     = "BNCVA"
	MethodSINARMAS  = "SINARMAS"
	MethodMANDIRIVA = "MANDIRIVA"
	MethodPERMATAVA = "PERMATAVA"
	MethodCIMBVA    = "CIMBVA"
	MethodDANAMON   = "DANAMON"
	MethodMUAMALAT  = "MUAMALAT"
	MethodBSIVA     = "BSIVA"
	MethodOCBC      = "OCBC"

	// E-Wallet Payment Methods
	MethodGOPAY     = "GOPAY"
	MethodDANA      = "DANA"
	MethodOVO       = "OVO"
	MethodSHOPEEPAY = "SHOPEEPAY"
	MethodLINKAJA   = "LINKAJA"

	// Retail Payment Methods
	MethodALFAMART  = "ALFAMART"
	MethodINDOMARET = "INDOMARET"
)

// FormatExpiredTime returns a formatted expiration time.
//
// Parses the Expired field string into a time.Time value.
// Returns an error if the format is invalid.
func (t *TransactionHistoryItem) FormatExpiredTime() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", t.Expired)
}

// FormatDateTime returns a formatted date time.
//
// Combines the Date and Time fields into a single time.Time value.
// Returns an error if the format is invalid.
func (t *TransactionHistoryItem) FormatDateTime() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", t.Date+" "+t.Time)
}

// IsExpired checks if transaction is expired.
//
// Returns true if the status equals StatusExpired.
func (s *TransactionStatusValue) IsExpired() bool {
	return *s == StatusExpired
}

// IsSuccess checks if transaction is successful.
//
// Returns true if the status equals StatusSuccess.
func (s *TransactionStatusValue) IsSuccess() bool {
	return *s == StatusSuccess
}

// IsPending checks if transaction is pending.
//
// Returns true if the status equals StatusPending.
func (s *TransactionStatusValue) IsPending() bool {
	return *s == StatusPending
}

// GetStringValue returns the string value of status.
//
// Returns the underlying string value of the TransactionStatusValue.
func (s *TransactionStatusValue) GetStringValue() string {
	return string(*s)
}
