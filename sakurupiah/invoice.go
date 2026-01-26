package sakurupiah

import (
	"fmt"
	"strconv"
)

// CreateInvoice creates a new payment invoice with the specified parameters.
//
// This is the main method for creating payment transactions. It validates all required fields,
// generates the required HMAC-SHA256 signature, and sends the request to the Sakurupiah API.
//
// The callback URL will receive payment status updates (pending, success, expired).
// The return URL is where customers will be redirected after payment.
//
// # Example
//
//	resp, err := client.CreateInvoice(sakurupiah.CreateInvoiceRequest{
//	    Method:       "QRIS",
//	    CustomerName: "John Doe",
//	    CustomerEmail: "john@example.com",
//	    CustomerPhone: "628123456789",
//	    Amount:       10000,
//	    MerchantFee:  int(sakurupiah.FeeTypeMerchant),
//	    MerchantRef:  "INV-2025-001",
//	    Expired:      24,
//	    CallbackURL:  "https://yourdomain.com/callback",
//	    ReturnURL:    "https://yourdomain.com/return",
//	})
//
// # Signature Generation
//
// The signature is automatically generated using the algorithm:
// HMAC-SHA256(apiID + method + merchantRef + amount, apiKey)
//
// # Required Fields
//   - Method: Payment channel code (e.g., "QRIS", "BCAVA")
//   - CustomerPhone: Customer phone number (Indonesian format: 628xxx or 08xxx)
//   - Amount: Transaction amount in IDR (> 0)
//   - MerchantRef: Your unique transaction reference
//   - CallbackURL: URL to receive payment status updates
//   - ReturnURL: URL to redirect after payment
//
// # Optional Fields
//   - CustomerName: Customer name
//   - CustomerEmail: Customer email
//   - MerchantFee: Who pays the fee (1=merchant, 2=customer, default=1)
//   - Expired: Payment expiration in hours (default depends on channel)
//   - Products: Array of product items
//
// # Return
//
// *CreateInvoiceResponse containing the transaction details including TrxID and checkout URL
func (c *Client) CreateInvoice(req CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	// Validate required fields
	if req.Method == "" {
		return nil, ErrMissingMethod
	}
	if req.CustomerPhone == "" {
		return nil, ErrInvalidPhone
	}
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if req.MerchantRef == "" {
		return nil, ErrMissingMerchantRef
	}
	if req.MerchantFee != int(FeeTypeMerchant) && req.MerchantFee != int(FeeTypeCustomer) {
		req.MerchantFee = int(FeeTypeMerchant) // Default to merchant pays
	}

	// Use default URLs if not provided
	callbackURL := req.CallbackURL
	if callbackURL == "" {
		callbackURL = c.callbackURL
	}
	if callbackURL == "" {
		return nil, fmt.Errorf("callback URL is required")
	}

	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = c.returnURL
	}
	if returnURL == "" {
		return nil, fmt.Errorf("return URL is required")
	}

	// Generate signature
	signature := c.GenerateSignature(req.Method, req.MerchantRef, req.Amount)

	// Build form data
	data := map[string]string{
		"api_id":        c.apiID,
		"method":        req.Method,
		"phone":         req.CustomerPhone,
		"amount":        strconv.FormatInt(req.Amount, 10),
		"merchant_fee":  strconv.Itoa(req.MerchantFee),
		"merchant_ref":  req.MerchantRef,
		"callback_url":  callbackURL,
		"return_url":    returnURL,
		"signature":     signature,
	}

	if req.CustomerName != "" {
		data["name"] = req.CustomerName
	}
	if req.CustomerEmail != "" {
		data["email"] = req.CustomerEmail
	}
	if req.Expired > 0 {
		data["expired"] = strconv.Itoa(req.Expired)
	}

	// Build arrays for products
	arrays := map[string][]string{}
	if len(req.Products) > 0 {
		for _, p := range req.Products {
			arrays["produk[]"] = append(arrays["produk[]"], p.Name)
			arrays["qty[]"] = append(arrays["qty[]"], strconv.Itoa(p.Qty))
			arrays["harga[]"] = append(arrays["harga[]"], strconv.FormatInt(p.Price, 10))
			if p.Size != "" {
				arrays["size[]"] = append(arrays["size[]"], p.Size)
			}
			if p.Note != "" {
				arrays["note[]"] = append(arrays["note[]"], p.Note)
			}
		}
	}

	var result CreateInvoiceResponse
	if err := c.doJSONRequestWithArray("create.php", data, arrays, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateInvoiceWithProducts creates a new invoice with product details.
//
// This is a convenience method that wraps CreateInvoice with product array support.
// Use this when you need to include line items in your invoice.
//
// # Example
//
//	products := []sakurupiah.Product{
//	    {Name: "T-Shirt", Qty: 1, Price: 50000, Size: "L", Note: "Blue"},
//	    {Name: "Pants", Qty: 2, Price: 75000, Size: "M", Note: "Black"},
//	}
//
//	resp, err := client.CreateInvoiceWithProducts(
//	    "QRIS",
//	    "628123456789",
//	    200000,
//	    "INV-001",
//	    products,
//	)
//
// # Parameters
//
//   - method: Payment channel code
//   - phone: Customer phone number
//   - amount: Total transaction amount in IDR
//   - merchantRef: Your unique transaction reference
//   - products: Array of product items
//
// # Return
//
// *CreateInvoiceResponse with transaction details
func (c *Client) CreateInvoiceWithProducts(
	method string,
	phone string,
	amount int64,
	merchantRef string,
	products []Product,
) (*CreateInvoiceResponse, error) {
	return c.CreateInvoice(CreateInvoiceRequest{
		Method:      method,
		CustomerPhone: phone,
		Amount:      amount,
		MerchantRef: merchantRef,
		Products:    products,
	})
}

// CreateInvoiceSimple creates a new invoice with minimal parameters.
//
// This is the simplest way to create an invoice when you don't need products
// or customer details beyond the phone number.
//
// # Example
//
//	resp, err := client.CreateInvoiceSimple(
//	    "QRIS",                       // Payment method
//	    "John Doe",                   // Customer name
//	    "628123456789",              // Customer phone
//	    10000,                        // Amount in IDR
//	    "INV-2025-001",              // Your unique reference
//	    "https://yourdomain.com/callback",
//	    "https://yourdomain.com/return",
//	)
//
// # Parameters
//
//   - method: Payment channel code
//   - name: Customer name
//   - phone: Customer phone number
//   - amount: Transaction amount in IDR
//   - merchantRef: Your unique transaction reference
//   - callbackURL: URL to receive payment status updates
//   - returnURL: URL to redirect after payment
//
// # Return
//
// *CreateInvoiceResponse with transaction details including TrxID and checkout URL
func (c *Client) CreateInvoiceSimple(
	method string,
	name string,
	phone string,
	amount int64,
	merchantRef string,
	callbackURL string,
	returnURL string,
) (*CreateInvoiceResponse, error) {
	return c.CreateInvoice(CreateInvoiceRequest{
		Method:       method,
		CustomerName: name,
		CustomerPhone: phone,
		Amount:       amount,
		MerchantRef:  merchantRef,
		CallbackURL:  callbackURL,
		ReturnURL:    returnURL,
	})
}
