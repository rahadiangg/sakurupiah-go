package main

import (
	"fmt"
	"log"
	"net/http"

	sakurupiah "github.com/rahadiangg/sakurupiah-go/sakurupiah"
)

func main() {
	// Create a new client
	client, err := sakurupiah.NewClient(sakurupiah.Config{
		APIID:     "YOUR_API_ID",
		APIKey:    "YOUR_API_KEY",
		IsSandbox: true, // Use true for testing, false for production
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

	// Example 3: Create a simple invoice
	invoiceResp, err := client.CreateInvoiceSimple(
		sakurupiah.MethodQRIS, // Payment method
		"John Doe",            // Customer name
		"628123456789",        // Customer phone
		10000,                 // Amount
		"INV-2025-001",        // Merchant reference
		"https://yourdomain.com/callback",
		"https://yourdomain.com/return",
	)
	if err != nil {
		log.Printf("Error creating invoice: %v", err)
	} else {
		fmt.Printf("\nInvoice Created:\n")
		if len(invoiceResp.Data) > 0 {
			fmt.Printf("Transaction ID: %s\n", invoiceResp.Data[0].TrxID)
			fmt.Printf("Payment Status: %s\n", invoiceResp.Data[0].PaymentStatus)
			fmt.Printf("Checkout URL: %s\n", invoiceResp.Data[0].CheckoutURL)
		}
	}

	// Example 4: Create invoice with products
	products := []sakurupiah.Product{
		{Name: "T-Shirt", Qty: 1, Price: 50000, Size: "L", Note: "Blue"},
		{Name: "Pants", Qty: 2, Price: 75000, Size: "M", Note: "Black"},
	}

	invoiceReq := sakurupiah.CreateInvoiceRequest{
		Method:        sakurupiah.MethodBCAVA,
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
	fmt.Printf("\nGenerated Signature: %s\n", signature)

	// Note: In production, you would start your HTTP server
	// log.Fatal(http.ListenAndServe(":8080", nil))
}
