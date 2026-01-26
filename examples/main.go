package main

import (
	"fmt"
	"log"
	"net/http"

	sakurupiah "github.com/rahadiangg/sakurupiah-go/sakurupiah"
)

func main() {
	// Create a new client with default payment method
	client, err := sakurupiah.NewClient(sakurupiah.Config{
		APIID:                "YOUR_API_ID",
		APIKey:               "YOUR_API_KEY",
		IsSandbox:            true,                  // Use true for testing, false for production
		DefaultPaymentMethod: sakurupiah.MethodQRIS, // Set default payment method
	})
	if err != nil {
		log.Fatal(err)
	}

	// Example 1: List all payment channels
	channels, err := client.ListPaymentChannels()
	if err != nil {
		log.Printf("Error listing payment channels: %v", err)
	} else {
		fmt.Printf("Found %d payment channels\n", len(channels.Data))
		for _, ch := range channels.Data {
			if ch.Status == "Aktif" {
				fmt.Printf("- %s: %s (Min: %s, Max: %s, Fee: %s)\n",
					ch.Code, ch.Name, ch.Min, ch.Max, ch.Fee)
			}
		}
	}

	// Example 2: Check balance
	balance, err := client.CheckBalance()
	if err != nil {
		log.Printf("Error checking balance: %v", err)
	} else {
		fmt.Printf("\nBalance Information:\n")
		fmt.Printf("Merchant: %s\n", balance.Data.MerchantName)
		fmt.Printf("Available Balance: %s\n", balance.Data.AvailableBalance)
	}

	// Example 3: Create a simple invoice using default payment method (QRIS)
	// When payment method is not specified, it uses the default from Config
	invoiceResp, err := client.CreateInvoiceSimple(
		"", // Empty method will use the default from Config
		"John Doe",
		"628123456789",
		10000,
		"INV-2025-001",
		"https://yourdomain.com/callback",
		"https://yourdomain.com/return",
	)
	if err != nil {
		log.Printf("Error creating invoice: %v", err)
	} else {
		fmt.Printf("\nInvoice Created (Default QRIS):\n")
		if len(invoiceResp.Data) > 0 {
			fmt.Printf("Transaction ID: %s\n", invoiceResp.Data[0].TrxID)
			fmt.Printf("Payment Status: %s\n", invoiceResp.Data[0].PaymentStatus)
			fmt.Printf("Checkout URL: %s\n", invoiceResp.Data[0].CheckoutURL)
		}
	}

	// Example 3b: Create invoice with explicit QRIS method
	invoiceResp2, err := client.CreateInvoiceSimple(
		sakurupiah.MethodQRIS2, // Using QRIS2 for higher limits
		"Jane Doe",
		"628123456789",
		5000000, // Higher amount for QRIS2
		"INV-2025-001b",
		"https://yourdomain.com/callback",
		"https://yourdomain.com/return",
	)
	if err != nil {
		log.Printf("Error creating QRIS2 invoice: %v", err)
	} else {
		fmt.Printf("\nInvoice Created (QRIS2):\n")
		if len(invoiceResp2.Data) > 0 {
			fmt.Printf("Transaction ID: %s\n", invoiceResp2.Data[0].TrxID)
			fmt.Printf("QR Code: %s\n", invoiceResp2.Data[0].QR)
		}
	}

	// Example 3c: Create invoice with Virtual Account (BCA)
	invoiceResp3, err := client.CreateInvoiceSimple(
		sakurupiah.MethodBCAVA, // BCA Virtual Account
		"Bob Smith",
		"628123456789",
		50000,
		"INV-2025-001c",
		"https://yourdomain.com/callback",
		"https://yourdomain.com/return",
	)
	if err != nil {
		log.Printf("Error creating VA invoice: %v", err)
	} else {
		fmt.Printf("\nInvoice Created (BCA VA):\n")
		if len(invoiceResp3.Data) > 0 {
			fmt.Printf("Transaction ID: %s\n", invoiceResp3.Data[0].TrxID)
			fmt.Printf("Payment Number: %s\n", invoiceResp3.Data[0].PaymentNo)
		}
	}

	// Example 3d: Create invoice with E-Wallet (DANA)
	invoiceResp4, err := client.CreateInvoiceSimple(
		sakurupiah.MethodDANA, // DANA E-Wallet
		"Alice Johnson",
		"628123456789",
		25000,
		"INV-2025-001d",
		"https://yourdomain.com/callback",
		"https://yourdomain.com/return",
	)
	if err != nil {
		log.Printf("Error creating DANA invoice: %v", err)
	} else {
		fmt.Printf("\nInvoice Created (DANA):\n")
		if len(invoiceResp4.Data) > 0 {
			fmt.Printf("Transaction ID: %s\n", invoiceResp4.Data[0].TrxID)
			fmt.Printf("Checkout URL: %s\n", invoiceResp4.Data[0].CheckoutURL)
		}
	}

	// Example 3e: Validate payment method before creating invoice
	if sakurupiah.IsValidPaymentMethod("QRIS") {
		fmt.Println("\nQRIS is a valid payment method")
	}
	// Check payment method category
	if sakurupiah.IsQRISMethod(sakurupiah.MethodQRIS) {
		fmt.Println("QRIS is a QRIS payment method")
	}
	if sakurupiah.IsVAMethod(sakurupiah.MethodBCAVA) {
		fmt.Println("BCAVA is a Virtual Account payment method")
	}
	if sakurupiah.IsEWalletMethod(sakurupiah.MethodDANA) {
		fmt.Println("DANA is an E-Wallet payment method")
	}

	// Example 3f: Get all payment methods by category
	qrisMethods := sakurupiah.GetPaymentMethodsByCategory("qris")
	fmt.Printf("\nQRIS Methods: %v\n", qrisMethods)
	vaMethods := sakurupiah.GetPaymentMethodsByCategory("va")
	fmt.Printf("VA Methods: %v\n", vaMethods)
	ewalletMethods := sakurupiah.GetPaymentMethodsByCategory("ewallet")
	fmt.Printf("E-Wallet Methods: %v\n", ewalletMethods)

	// Example 4: Create invoice with products (using BRI Virtual Account)
	products := []sakurupiah.Product{
		{Name: "T-Shirt", Qty: 1, Price: 50000, Size: "L", Note: "Blue"},
		{Name: "Pants", Qty: 2, Price: 75000, Size: "M", Note: "Black"},
	}

	invoiceReq := sakurupiah.CreateInvoiceRequest{
		Method:        sakurupiah.MethodBRIVA, // Using BRI Virtual Account
		CustomerName:  "Jane Smith",
		CustomerEmail: "jane@example.com",
		CustomerPhone: "628987654321",
		Amount:        200000,
		MerchantFee:   int(sakurupiah.FeeTypeMerchant),
		MerchantRef:   "INV-2025-002",
		Expired:       24,
		Products:      products,
		CallbackURL:   "https://yourdomain.com/callback",
		ReturnURL:     "https://yourdomain.com/return",
	}

	invoiceWithProducts, err := client.CreateInvoice(invoiceReq)
	if err != nil {
		log.Printf("Error creating invoice with products: %v", err)
	} else {
		fmt.Printf("\nInvoice with Products Created:\n")
		if len(invoiceWithProducts.Data) > 0 {
			fmt.Printf("Transaction ID: %s\n", invoiceWithProducts.Data[0].TrxID)
			fmt.Printf("Payment Number: %s\n", invoiceWithProducts.Data[0].PaymentNo)
		}
		fmt.Printf("Products: %d\n", len(invoiceWithProducts.Product))
	}

	// Example 5: Get transaction history
	history, err := client.GetTransactionsByStatus("berhasil")
	if err != nil {
		log.Printf("Error getting transaction history: %v", err)
	} else {
		fmt.Printf("\nSuccessful Transactions: %d\n", len(history.Data))
		for _, trx := range history.Data {
			fmt.Printf("- %s: %s (%s)\n", trx.TrxID, trx.Amount, trx.Status)
		}
	}

	// Example 6: Get transactions by date range
	dateHistory, err := client.GetTransactionsByDateRange("2025-01-01", "2025-01-31")
	if err != nil {
		log.Printf("Error getting transactions by date: %v", err)
	} else {
		fmt.Printf("\nTransactions in January 2025: %d\n", len(dateHistory.Data))
	}

	// Example 7: Check transaction status
	trxID := "TRX123" // Replace with actual transaction ID
	status, err := client.GetTransactionStatus(trxID)
	if err != nil {
		log.Printf("Error checking transaction status: %v", err)
	} else {
		fmt.Printf("\nTransaction Status: %s\n", status.Data[0].Status)
	}

	// Example 8: Set up callback handler
	// In a real application, you would register this with your HTTP server
	callbackHandler := client.NewCallbackHandler(func(callback *sakurupiah.CallbackRequest) error {
		fmt.Printf("\nCallback Received:\n")
		fmt.Printf("Transaction ID: %s\n", callback.TrxID)
		fmt.Printf("Merchant Ref: %s\n", callback.MerchantRef)
		fmt.Printf("Status: %s\n", callback.Status)

		// Update your database based on payment status
		switch callback.Status {
		case sakurupiah.StatusSuccess:
			fmt.Println("Payment successful - update order status")
			// TODO: Update your database
		case sakurupiah.StatusExpired:
			fmt.Println("Payment expired - handle expiration")
			// TODO: Handle expiration
		case sakurupiah.StatusPending:
			fmt.Println("Payment still pending")
			// TODO: Handle pending status
		}

		return nil
	})

	// Register the handler
	http.HandleFunc("/callback", callbackHandler)
	fmt.Println("\nCallback handler registered at /callback")

	// Example 9: Using callback handler builder for different statuses
	builder := sakurupiah.NewCallbackHandlerBuilder(client).
		OnSuccess(func(callback *sakurupiah.CallbackRequest) error {
			fmt.Printf("Payment successful for TrxID: %s\n", callback.TrxID)
			// TODO: Process successful payment
			return nil
		}).
		OnExpired(func(callback *sakurupiah.CallbackRequest) error {
			fmt.Printf("Payment expired for TrxID: %s\n", callback.TrxID)
			// TODO: Handle expired payment
			return nil
		}).
		OnPending(func(callback *sakurupiah.CallbackRequest) error {
			fmt.Printf("Payment pending for TrxID: %s\n", callback.TrxID)
			// TODO: Handle pending payment
			return nil
		})

	advancedHandler := builder.Build()
	http.HandleFunc("/callback-advanced", advancedHandler)

	// Example 10: Generate signature manually (for testing)
	signature := client.GenerateSignature(sakurupiah.MethodQRIS, "TEST-REF-123", 10000)
	fmt.Printf("\nGenerated Signature for QRIS: %s\n", signature)

	// Example 11: Change default payment method at runtime
	client.SetDefaultPaymentMethod(sakurupiah.MethodOVO)
	fmt.Printf("Default payment method changed to: %s\n", client.GetDefaultPaymentMethod())

	// Now this invoice will use OVO as the payment method
	invoiceResp5, err := client.CreateInvoiceSimple(
		"", // Will use OVO (the new default)
		"Customer Name",
		"628123456789",
		15000,
		"INV-2025-001e",
		"https://yourdomain.com/callback",
		"https://yourdomain.com/return",
	)
	if err != nil {
		log.Printf("Error creating OVO invoice: %v", err)
	} else {
		fmt.Printf("\nInvoice Created (OVO - Default):\n")
		if len(invoiceResp5.Data) > 0 {
			fmt.Printf("Transaction ID: %s\n", invoiceResp5.Data[0].TrxID)
		}
	}

	// Note: In production, you would start your HTTP server
	// log.Fatal(http.ListenAndServe(":8080", nil))
}
