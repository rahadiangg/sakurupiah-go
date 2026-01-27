# Sakurupiah Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/rahadiangg/sakurupiah-go.svg)](https://pkg.go.dev/github.com/rahadiangg/sakurupiah-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/rahadiangg/sakurupiah-go)](https://goreportcard.com/report/github.com/rahadiangg/sakurupiah-go)
[![Testing](https://github.com/rahadiangg/sakurupiah-go/actions/workflows/test.yml/badge.svg)](https://github.com/rahadiangg/sakurupiah-go/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-%2300ADD8.svg)](https://golang.org)

> **⚠️ Unofficial SDK** - This is an unofficial Go SDK for the [Sakurupiah Payment Gateway](https://sakurupiah.id). It is not officially maintained or endorsed by Sakurupiah.

## Features

- :moneybag: Create payment invoices with single or multiple products
- :list: List available payment channels with real-time status
- :balance_scale: Check merchant balance
- :mag: Query transaction history with advanced filters
- :eye: Check transaction status
- :lock: Handle payment callbacks with HMAC-SHA256 signature verification
- :globe: Support for both production and sandbox environments

## Installation

```bash
go get github.com/rahadiangg/sakurupiah-go/sakurupiah
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    sakurupiah "github.com/rahadiangg/sakurupiah-go/sakurupiah"
)

func main() {
    // Initialize client
    client, err := sakurupiah.NewClient(sakurupiah.Config{
        APIID:    "YOUR_API_ID",
        APIKey:   "YOUR_API_KEY",
        IsSandbox: true, // Set to false for production
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create a simple invoice
    resp, err := client.CreateInvoiceSimple(
        sakurupiah.MethodQRIS,        // Payment method (use constants for type safety)
        "John Doe",                   // Customer name
        "628123456789",              // Customer phone
        10000,                        // Amount in IDR
        "INV-2025-001",              // Your unique reference
        "https://yourdomain.com/callback",
        "https://yourdomain.com/return",
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Transaction ID: %s\n", resp.Data[0].TrxID)
    fmt.Printf("Checkout URL: %s\n", resp.Data[0].CheckoutURL)
}
```

## Documentation

Full documentation is available at [https://pkg.go.dev/github.com/rahadiangg/sakurupiah-go](https://pkg.go.dev/github.com/rahadiangg/sakurupiah-go)

### Table of Contents

- [Initialization](#initialization)
- [Creating Invoices](#creating-invoices)
- [Payment Channels](#payment-channels)
- [Transaction History](#transaction-history)
- [Callback Handling](#callback-handling)
- [Configuration](#configuration)
- [Error Handling](#error-handling)
- [Testing](#testing)

### Initialization

```go
client, err := sakurupiah.NewClient(sakurupiah.Config{
    APIID:              "YOUR_API_ID",
    APIKey:             "YOUR_API_KEY",
    IsSandbox:           true,
    Timeout:            30 * time.Second,
    DefaultCallbackURL:  "https://yourdomain.com/callback",
    DefaultReturnURL:    "https://yourdomain.com/return",
})
```

### Creating Invoices

#### Simple Invoice

```go
resp, err := client.CreateInvoiceSimple(
    sakurupiah.MethodQRIS,     // Payment method
    "John Doe",                // Customer name
    "628123456789",           // Customer phone
    10000,                     // Amount in IDR
    "INV-2025-001",           // Your unique reference
    "https://yourdomain.com/callback",
    "https://yourdomain.com/return",
)
```

#### Invoice with Products

```go
products := []sakurupiah.Product{
    {Name: "T-Shirt", Qty: 1, Price: 50000, Size: "L", Note: "Blue"},
    {Name: "Pants", Qty: 2, Price: 75000, Size: "M", Note: "Black"},
}

req := sakurupiah.CreateInvoiceRequest{
    Method:       sakurupiah.MethodQRIS,
    CustomerName: "John Doe",
    CustomerEmail: "john@example.com",
    CustomerPhone: "628123456789",
    Amount:       200000,
    MerchantFee:  int(sakurupiah.FeeTypeMerchant), // or FeeTypeCustomer
    MerchantRef:  "INV-2025-001",
    Expired:      24, // 24 hours
    Products:     products,
    CallbackURL:  "https://yourdomain.com/callback",
    ReturnURL:    "https://yourdomain.com/return",
}

resp, err := client.CreateInvoice(req)
```

#### Configuring Merchant Fee

```go
// Option 1: Merchant pays the fee (default)
req.MerchantFee = int(sakurupiah.FeeTypeMerchant) // = 1

// Option 2: Customer pays the fee
req.MerchantFee = int(sakurupiah.FeeTypeCustomer) // = 2
```

#### Available Payment Methods

The SDK provides type-safe constants for all supported payment methods:

```go
// QRIS Methods
sakurupiah.MethodQRIS   // "QRIS"
sakurupiah.MethodQRIS2  // "QRIS2"
sakurupiah.MethodQRISMU // "QRISMU"
sakurupiah.MethodQRISC  // "QRISC"

// Virtual Accounts
sakurupiah.MethodBCAVA     // "BCAVA"
sakurupiah.MethodBRIVA     // "BRIVA"
sakurupiah.MethodBNIVA     // "BNIVA"
// ... and more

// E-Wallets
sakurupiah.MethodGOPAY     // "GOPAY"
sakurupiah.MethodDANA      // "DANA"
sakurupiah.MethodOVO       // "OVO"
sakurupiah.MethodSHOPEEPAY // "ShopeePay"
sakurupiah.MethodLINKAJA   // "LINKAJA"

// Retail
sakurupiah.MethodALFAMART  // "ALFAMART"
sakurupiah.MethodINDOMARET // "INDOMARET"
```

### Payment Channels

```go
channels, err := client.ListPaymentChannels()
if err != nil {
    log.Fatal(err)
}

for _, ch := range channels.Data {
    if ch.Status == "Aktif" {
        fmt.Printf("%s: %s (Min: %s, Max: %s, Fee: %s)\n",
            ch.Code, ch.Name, ch.Min, ch.Max, ch.Fee)
    }
}
```

### Check Balance

```go
balance, err := client.CheckBalance()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Merchant: %s\n", balance.Data.MerchantName)
fmt.Printf("Available Balance: %s\n", balance.Data.AvailableBalance)
```

### Transaction History

```go
// Get all successful transactions
history, err := client.GetTransactionsByStatus("berhasil")

// Get transactions by date range
history, err = client.GetTransactionsByDateRange("2025-01-01", "2025-01-31")

// Get transaction by ID
history, err = client.GetTransactionByTrxID("TRX123")

// Get transactions by merchant reference
history, err = client.GetTransactionsByMerchantRef("INV-2025-001")

// Advanced filtering
history, err = client.GetTransactionHistory(sakurupiah.TransactionHistoryRequest{
    Status:      "pending",
    PaymentCode: sakurupiah.MethodQRIS,
    StartDate:   "2025-01-01",
    EndDate:     "2025-01-31",
})
```

### Check Transaction Status

```go
status, err := client.GetTransactionStatus("TRX123")
if err != nil {
    log.Fatal(err)
}

switch status.Data[0].Status {
case sakurupiah.StatusSuccess:
    fmt.Println("Payment successful!")
case sakurupiah.StatusPending:
    fmt.Println("Payment pending")
case sakurupiah.StatusExpired:
    fmt.Println("Payment expired")
}
```

## Callback Handling

The SDK provides secure callback handling with signature verification.

### Simple Handler

```go
handler := client.NewCallbackHandler(func(callback *sakurupiah.CallbackRequest) error {
    fmt.Printf("TrxID: %s, Status: %s\n", callback.TrxID, callback.Status)

    switch callback.Status {
    case sakurupiah.StatusSuccess:
        // Update order status to paid
        // TODO: Update your database
    case sakurupiah.StatusExpired:
        // Handle expired payment
    case sakurupiah.StatusPending:
        // Payment still pending
    }

    return nil
})

http.HandleFunc("/callback", handler)
log.Fatal(http.ListenAndServe(":8080", nil))
```

### Advanced Handler with Builder

```go
builder := sakurupiah.NewCallbackHandlerBuilder(client).
    OnSuccess(func(callback *sakurupiah.CallbackRequest) error {
        fmt.Printf("Payment successful: %s\n", callback.TrxID)
        // Process successful payment
        return nil
    }).
    OnExpired(func(callback *sakurupiah.CallbackRequest) error {
        fmt.Printf("Payment expired: %s\n", callback.TrxID)
        // Handle expiration
        return nil
    }).
    OnPending(func(callback *sakurupiah.CallbackRequest) error {
        fmt.Printf("Payment pending: %s\n", callback.TrxID)
        // Handle pending status
        return nil
    })

handler := builder.Build()
http.HandleFunc("/callback", handler)
```

### Manual Verification

```go
func callbackHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    callback, err := client.VerifyAndParseCallback(r.Header, body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Process callback
    fmt.Printf("Payment %s is %s\n", callback.TrxID, callback.Status)

    // Send response
    sakurupiah.SendCallbackResponse(w, sakurupiah.CallbackResponse{
        Success: true,
        Message: "Payment processed",
    })
}
```

## Payment Methods

| Code | Name | Type | Min Amount | Fee |
|------|------|------|------------|-----|
| QRIS | QRIS | DIRECT | Rp 500 | 0.7% |
| QRIS2 | QRIS2 | DIRECT | Rp 100 | 0.9% |
| BCAVA | BCA Virtual Account | DIRECT | Rp 10.000 | Rp 4.900 |
| BRIVA | BRI Virtual Account | DIRECT | Rp 10.000 | Rp 3.500 |
| BNIVA | BNI Virtual Account | DIRECT | Rp 10.000 | Rp 3.500 |
| MANDIRIVA | Mandiri Virtual Account | DIRECT | Rp 10.000 | Rp 3.500 |
| GOPAY | GoPay | REDIRECT | Rp 500 | 3% |
| DANA | DANA | REDIRECT | Rp 1.000 | 3% |
| OVO | OVO | REDIRECT | Rp 1.000 | 3% |
| SHOPEEPAY | ShopeePay | REDIRECT | Rp 1.000 | 3% |
| ALFAMART | Alfamart | DIRECT | Rp 10.000 | Rp 3.000 |

*For a complete list of payment methods, use `ListPaymentChannels()`*

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `APIID` | string | *Required* | Your Sakurupiah API ID |
| `APIKey` | string | *Required* | Your Sakurupiah API Key |
| `IsSandbox` | bool | `false` | Use sandbox environment |
| `Timeout` | time.Duration | `30s` | HTTP request timeout |
| `HTTPClient` | `*http.Client` | `nil` | Custom HTTP client |
| `DefaultCallbackURL` | string | `""` | Default callback URL for invoices |
| `DefaultReturnURL` | string | `""` | Default return URL for invoices |

## Error Handling

```go
resp, err := client.CreateInvoice(req)
if err != nil {
    if apiErr, ok := err.(*sakurupiah.APIError); ok {
        // API returned an error
        log.Printf("API Error (status %s): %s", apiErr.Status, apiErr.Message)
        return
    }
    // Other error (network, validation, etc.)
    log.Fatal(err)
}
```

### Common Errors

| Error | Description |
|-------|-------------|
| `ErrMissingAPIID` | API ID is required |
| `ErrMissingAPIKey` | API Key is required |
| `ErrInvalidAmount` | Invalid amount (must be > 0) |
| `ErrInvalidPhone` | Invalid phone number format |
| `ErrMissingMerchantRef` | Merchant reference is required |
| `ErrMissingMethod` | Payment method is required |
| `ErrInvalidSignature` | Invalid signature |

## Testing

### Run Unit Tests

```bash
go test ./...
```

### Run Integration Tests

```bash
go test -tags=integration ./...
```

### Run with Coverage

```bash
go test -cover ./...
```

### Run Specific Test

```bash
go test -run TestCreateInvoice ./...
```

## Examples

See the [examples](examples) directory for more usage examples:

- [Basic Usage](examples/main.go) - Complete example with all features
- More examples coming soon...

## API Documentation

For complete API documentation from Sakurupiah, visit:
https://sakurupiah.id/developers/api-dokumentasi

## Support

- 📧 Email: [support@sakurupiah.id](mailto:support@sakurupiah.id)
- 💬 WhatsApp: [+62 821 3301 4951](https://wa.me/6282133014951)
- 📚 Documentation: [https://sakurupiah.id/developers](https://sakurupiah.id/developers)

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Disclaimer

**This is an UNOFFICIAL SDK** and is not affiliated with, endorsed by, or officially supported by Sakurupiah. It is a community-maintained project for integration purposes.

Please refer to the [official Sakurupiah documentation](https://sakurupiah.id/developers) for the most up-to-date API information and official support channels.
