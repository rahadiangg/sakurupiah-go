package sakurupiah

// ListPaymentChannels retrieves all available payment channels from Sakurupiah.
//
// This includes both active and inactive channels with their current status,
// transaction limits (min/max amounts), fee structures, and payment instructions.
//
// Use this method to:
//   - Display payment options to your customers
//   - Check if a payment method is currently available
//   - Get current fee rates and limits
//   - Display payment instructions to customers
//
// # Example
//
//	channels, err := client.ListPaymentChannels()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, ch := range channels.Data {
//	    if ch.Status == "Aktif" {
//	        fmt.Printf("%s: %s (Min: %s, Max: %s, Fee: %s)\n",
//	            ch.Code, ch.Name, ch.Min, ch.Max, ch.Fee)
//	    }
//	}
//
// # Channel Information
//
// Each channel includes:
//   - Code: Payment method code for API requests
//   - Name: Display name
//   - Status: "Aktif" (active) or "Offline" (inactive)
//   - Min/Max: Transaction amount limits
//   - Fee: Transaction fee (amount or percentage)
//   - Type: "DIRECT" or "REDIRECT"
//   - Guide: Payment instructions
//
// # Return
//
// *ListPaymentChannelsResponse containing all available payment channels
func (c *Client) ListPaymentChannels() (*ListPaymentChannelsResponse, error) {
	data := map[string]string{
		"api_id": c.apiID,
		"method": "list",
	}

	var result ListPaymentChannelsResponse
	if err := c.doJSONRequest("list-payment.php", buildFormData(data), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CheckBalance retrieves the merchant's balance information from Sakurupiah.
//
// Returns two balance types:
//   - Balance: Pending settlement balance (transactions awaiting settlement)
//   - Available Balance: Ready-to-use balance that has been settled
//
// Use this to:
//   - Display current balance to merchants
//   - Monitor fund availability
//   - Track settlement status
//
// # Example
//
//	balance, err := client.CheckBalance()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Merchant: %s\n", balance.Data.MerchantName)
//	fmt.Printf("Pending Balance: %s\n", balance.Data.Balance)
//	fmt.Printf("Available Balance: %s\n", balance.Data.AvailableBalance)
//
// # Balance Information
//
// - Balance: Funds from successful transactions that are pending settlement
// - AvailableBalance: Funds that have been settled and can be withdrawn
//
// Settlement times vary by payment method (typically H+1 to H+3).
//
// # Return
//
// *CheckBalanceResponse containing merchant name and balance information
func (c *Client) CheckBalance() (*CheckBalanceResponse, error) {
	data := map[string]string{
		"api_id": c.apiID,
		"method": "balance",
	}

	var result CheckBalanceResponse
	if err := c.doJSONRequest("check_balance.php", buildFormData(data), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
