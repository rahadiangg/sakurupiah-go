# Changelog

> **Note**: This is an UNOFFICIAL SDK and is not affiliated with or endorsed by Sakurupiah.

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-26

### Added
- Initial release of Sakurupiah Go SDK
- Invoice creation with `CreateInvoice`, `CreateInvoiceSimple`, and `CreateInvoiceWithProducts`
- Payment channels listing with `ListPaymentChannels`
- Balance checking with `CheckBalance`
- Transaction history queries with filters (`GetTransactionHistory`, `GetTransactionsByStatus`, `GetTransactionsByPaymentCode`, etc.)
- Transaction status checking with `GetTransactionStatus`
- Secure callback handling with signature verification (`NewCallbackHandler`, `CallbackHandlerBuilder`)
- HMAC-SHA256 signature generation and verification
- Support for both production and sandbox environments
- Flexible JSON types for handling mixed string/number API responses
- Comprehensive unit and integration tests
- Full godoc documentation and examples

### Payment Methods Supported
- **QRIS**: QRIS, QRIS2, QRISM, QRISC
- **Virtual Accounts**: BCAVA, BRIVA, BNIVA, BAGVA, BNCVA, SINARMAS, MANDIRIVA, PERMATAVA, CIMBVA, DANAMON, MUAMALAT, BSIVA, OCBC
- **E-Wallets**: GOPAY, DANA, OVO, ShopeePay, LinkAja
- **Retail**: Alfamart, Indomaret

[1.0.0]: https://github.com/rahadiangg/sakurupiah-go/releases/tag/v1.0.0
[Unreleased]: https://github.com/rahadiangg/sakurupiah-go/compare/v1.0.0...HEAD
