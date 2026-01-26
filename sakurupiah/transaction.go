package sakurupiah

import (
	"strconv"
)

// GetTransactionHistory retrieves transaction history with optional filters.
//
// This method allows you to query transactions with various filters including
// status, date range, payment channel, merchant reference, and more.
// Leave filters empty to get all transactions.
//
// # Example
//
//	history, err := client.GetTransactionHistory(sakurupiah.TransactionHistoryRequest{
//	    Status:    "berhasil",
//	    StartDate: "2025-01-01",
//	    EndDate:   "2025-01-31",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, trx := range history.Data {
//	    fmt.Printf("%s: %s - %s\n", trx.TrxID, trx.MerchantRef, trx.Status)
//	}
//
// # Filters
//
//   - MerchantFilter: Filter by merchant (1 = current only, 0 or empty = all)
//   - PaymentCode: Filter by payment channel code (e.g., "QRIS", "BCAVA")
//   - TrxID: Filter by specific transaction ID
//   - MerchantRef: Filter by your merchant reference
//   - Status: Filter by status ("pending", "berhasil", "expired")
//   - StartDate/EndDate: Filter by date range (format: "YYYY-MM-DD")
//
// # Return
//
// *TransactionHistoryResponse containing filtered transaction list
func (c *Client) GetTransactionHistory(req TransactionHistoryRequest) (*TransactionHistoryResponse, error) {
	data := map[string]string{
		"api_id": c.apiID,
		"method": "transaction",
	}

	if req.MerchantFilter > 0 {
		data["mechant"] = strconv.Itoa(req.MerchantFilter)
	}
	if req.PaymentCode != "" {
		data["payment_kode"] = req.PaymentCode
	}
	if req.TrxID != "" {
		data["trx_id"] = req.TrxID
	}
	if req.MerchantRef != "" {
		data["merchant_ref"] = req.MerchantRef
	}
	if req.Status != "" {
		data["status"] = req.Status
	}
	if req.StartDate != "" {
		data["tanggal_awal"] = req.StartDate
	}
	if req.EndDate != "" {
		data["tanggal_akhir"] = req.EndDate
	}

	var result TransactionHistoryResponse
	if err := c.doJSONRequest("transaction.php", buildFormData(data), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetAllTransactions retrieves all transactions without filters.
//
// This is a convenience method that returns all transactions for your account.
// Use GetTransactionHistory with filters for more specific queries.
//
// # Example
//
//	history, err := client.GetAllTransactions()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Found %d transactions\n", len(history.Data))
//
// # Return
//
// *TransactionHistoryResponse containing all transactions
func (c *Client) GetAllTransactions() (*TransactionHistoryResponse, error) {
	return c.GetTransactionHistory(TransactionHistoryRequest{})
}

// GetTransactionsByStatus retrieves transactions filtered by status.
//
// This is a convenience method for filtering transactions by payment status.
//
// # Valid Statuses
//
//   - "pending": Transactions awaiting payment
//   - "berhasil": Successfully completed transactions
//   - "expired": Transactions that expired without payment
//
// # Example
//
//	successful, err := client.GetTransactionsByStatus("berhasil")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, trx := range successful.Data {
//	    fmt.Printf("%s: %s\n", trx.TrxID, trx.Amount)
//	}
//
// # Parameters
//
//   - status: One of "pending", "berhasil", or "expired"
//
// # Return
//
// *TransactionHistoryResponse containing filtered transactions
func (c *Client) GetTransactionsByStatus(status string) (*TransactionHistoryResponse, error) {
	return c.GetTransactionHistory(TransactionHistoryRequest{
		Status: status,
	})
}

// GetTransactionsByPaymentCode retrieves transactions filtered by payment channel code.
//
// This is a convenience method for filtering transactions by payment method.
//
// # Example
//
//	qrisTrx, err := client.GetTransactionsByPaymentCode("QRIS")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Parameters
//
//   - paymentCode: Payment channel code (e.g., "QRIS", "BCAVA", "DANA")
//
// # Return
//
// *TransactionHistoryResponse containing filtered transactions
func (c *Client) GetTransactionsByPaymentCode(paymentCode string) (*TransactionHistoryResponse, error) {
	return c.GetTransactionHistory(TransactionHistoryRequest{
		PaymentCode: paymentCode,
	})
}

// GetTransactionsByMerchantRef retrieves transactions filtered by merchant reference.
//
// This is useful for looking up the status of a specific transaction using
// your unique reference that was provided when creating the invoice.
//
// # Example
//
//	trx, err := client.GetTransactionsByMerchantRef("INV-2025-001")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if len(trx.Data) > 0 {
//	    fmt.Printf("Status: %s\n", trx.Data[0].Status)
//	}
//
// # Parameters
//
//   - merchantRef: Your unique transaction reference
//
// # Return
//
// *TransactionHistoryResponse containing matching transactions
func (c *Client) GetTransactionsByMerchantRef(merchantRef string) (*TransactionHistoryResponse, error) {
	return c.GetTransactionHistory(TransactionHistoryRequest{
		MerchantRef: merchantRef,
	})
}

// GetTransactionsByDateRange retrieves transactions within a date range.
//
// This is useful for generating reports and reconciling transactions.
//
// # Example
//
//	january, err := client.GetTransactionsByDateRange("2025-01-01", "2025-01-31")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Parameters
//
//   - startDate: Start date in "YYYY-MM-DD" format
//   - endDate: End date in "YYYY-MM-DD" format
//
// # Return
//
// *TransactionHistoryResponse containing transactions in the date range
func (c *Client) GetTransactionsByDateRange(startDate, endDate string) (*TransactionHistoryResponse, error) {
	return c.GetTransactionHistory(TransactionHistoryRequest{
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetTransactionByTrxID retrieves a specific transaction by its transaction ID.
//
// Use this to get full details of a transaction using the Sakurupiah transaction ID.
//
// # Example
//
//	trx, err := client.GetTransactionByTrxID("TRX-12345")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Parameters
//
//   - trxID: The Sakurupiah transaction ID
//
// # Return
//
// *TransactionHistoryResponse containing the transaction details
func (c *Client) GetTransactionByTrxID(trxID string) (*TransactionHistoryResponse, error) {
	return c.GetTransactionHistory(TransactionHistoryRequest{
		TrxID: trxID,
	})
}

// GetTransactionStatus checks the status of a specific transaction.
//
// This is the recommended method for checking the current status of a transaction.
// It returns only the status information, which is more efficient than
// retrieving the full transaction history.
//
// # Example
//
//	status, err := client.GetTransactionStatus("TRX-12345")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if len(status.Data) > 0 {
//	    fmt.Printf("Transaction status: %s\n", status.Data[0].Status)
//	}
//
// # Parameters
//
//   - trxID: The Sakurupiah transaction ID
//
// # Return
//
// *TransactionStatusResponse containing the transaction status
func (c *Client) GetTransactionStatus(trxID string) (*TransactionStatusResponse, error) {
	if trxID == "" {
		return nil, ErrMissingMerchantRef
	}

	data := map[string]string{
		"api_id": c.apiID,
		"method": "status",
		"trx_id": trxID,
	}

	var result TransactionStatusResponse
	if err := c.doJSONRequest("status-transaction.php", buildFormData(data), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
