# Contributing to Sakurupiah Go SDK

Thank you for your interest in contributing to the Sakurupiah Go SDK!

## Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/rahadiangg/sakurupiah-go.git
   cd sakurupiah-go
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run tests**
   ```bash
   # Run unit tests only
   go test -short ./...

   # Run all tests including integration tests
   go test -tags=integration ./...

   # Run with coverage
   go test -cover ./...
   ```

## Code Style

- Follow standard Go conventions defined in [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting code
- Run `golint` or `staticcheck` to check for issues
- Add godoc comments for exported functions, types, and constants

## Testing

- Unit tests should not require external API calls
- Use build tags for integration tests: `//go:build integration`
- Integration tests should use the sandbox environment only
- Ensure all tests pass before submitting a pull request

## Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Pull Request Guidelines

- Provide a clear description of the changes
- Reference any related issues
- Ensure all tests pass
- Update documentation if needed
- Follow the existing code style

## Reporting Issues

When reporting issues, please provide:

- Go version (`go version`)
- SDK version (git tag or commit hash)
- Detailed description of the problem
- Code snippets to reproduce the issue
- Error messages and stack traces

## Development Environment

- Go 1.21 or higher
- Access to Sakurupiah sandbox account for integration testing

## Documentation

- Update godoc comments for any public API changes
- Add examples for new features
- Keep CHANGELOG.md updated

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
