# Contributing Guidelines

## Development Setup
1. Fork and clone the repository
2. Install Go 1.21+ and AWS CLI
3. Run `go mod download` to install dependencies
4. Use `make run-server` for local development

## Code Standards
- Follow Go conventions and use `gofmt`
- Run `make lint` before submitting
- Add tests for new functionality
- Update documentation as needed

## Pull Request Process
1. Create a feature branch from main
2. Make your changes with appropriate tests
3. Run `make test` and `make lint`
4. Submit PR with clear description
5. Address review feedback

## Project Structure
Follow the established patterns:
- Core logic in `pkg/` packages
- Applications in `cmd/` directories
- Tests alongside source files
- Mock implementations for testing