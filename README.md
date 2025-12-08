# GLAD Stack - Go, Lambda, ApiGateway, DynamoDB

A comprehensive serverless API platform built with Go, demonstrating modern cloud-native architecture using AWS serverless technologies and production ready API.

## What is GLAD?

**GLAD** stands for:
- **G**o - Modern, efficient programming language with excellent performance and concurrency
- **L**ambda - AWS serverless compute platform for running code without managing servers
- **A**piGateway - AWS managed API gateway service for creating, deploying, and managing REST APIs
- **D**ynamoDB - AWS NoSQL database(Single Table Design) service providing fast and predictable performance with seamless scalability

This project showcases how these four technologies work together to create a production-ready, scalable, and cost-effective serverless API platform that can handle millions of requests while maintaining low latency and high availability.

[![Tests](https://img.shields.io/badge/tests-passing-brightgreen)]()
[![Go Version](https://img.shields.io/badge/go-1.24.0-blue)]()
[![Production Ready](https://img.shields.io/badge/status-production%20ready-success)]()

## Features

- **RESTful API** with user authentication and management
- **JWT Authentication** with token-based authorization
- **Serverless Architecture** using AWS Lambda + API Gateway
- **DynamoDB Integration** for data persistence
- **Clean Architecture** with layered design (Handler → Service → Repository)
- **Comprehensive Testing** with unit and integration tests
- **Structured Logging** using Go's slog package
- **Infrastructure as Code** with AWS CDK

## Project Structure

```
glad/
├── cmd/
│   └── app/                        # Lambda application
│       ├── main.go                 # Lambda entry point
│       ├── integration_test.go     # Integration tests
│       ├── testdata/               # Test data files
│       │   └── sample_request.json # Sample API requests
│       └── internal/               # App-specific code
│           ├── database/           # Repository implementations
│           ├── dto/                # Request/Response DTOs
│           ├── errors/             # App-specific errors
│           ├── handler/            # HTTP handlers (thin layer)
│           ├── models/             # Domain models
│           ├── router/             # Router abstraction
│           ├── service/            # Business logic
│           └── validation/         # Input validation
├── internal/                       # Shared private code
│   └── errors/                     # Domain error definitions
├── pkg/                            # Shared public packages
│   ├── auth/                       # JWT token service
│   ├── config/                     # Configuration management
│   ├── errors/                     # Core error utilities
│   ├── logger/                     # Structured logging
│   └── middleware/                 # HTTP middleware
├── deployments/
│   └── app/                        # AWS CDK infrastructure
│       ├── cdk.json                # CDK configuration
│       └── cdk.out/                # CDK build output
├── scripts/                        # Utility scripts
├── site/                           # Documentation/website
├── .github/
│   ├── workflows/
│   │   └── ci.yml                  # GitHub Actions CI
│   └── dependabot.yml              # Dependency updates
├── Taskfile.yml                    # Task runner configuration
├── .golangci.yml                   # Go linter configuration
└── README.md                       # This file
```

## Architecture

```
Request → Router → Middleware → Handler → Service → Repository → Database
                                   ↓
                               Validation
```

### Layers

1. **Router** - Route matching and middleware chaining
2. **Middleware** - JWT validation, logging, CORS
3. **Handler** - HTTP layer (JSON marshaling/unmarshaling)
4. **Service** - Business logic and validation
5. **Repository** - Data access abstraction
6. **Database** - DynamoDB or Mock implementation

### Design Patterns

- **Layered Architecture** - Clear separation of concerns
- **Repository Pattern** - Interface-based data access
- **Dependency Injection** - Config-based initialization
- **DTO Pattern** - Separate request/response types from domain models
- **Service Layer** - Business logic isolated from HTTP concerns

## API Endpoints

| Method | Path       | Auth | Description                  |
|--------|------------|------|------------------------------|
| POST   | /register  | No   | User registration            |
| POST   | /login     | No   | Authentication (returns JWT) |
| GET    | /protected | JWT  | Protected resource demo      |
| PUT    | /user      | JWT  | Update user profile          |
| GET    | /users     | JWT  | List all users               |

## Getting Started

### Prerequisites

- Go 1.21+ (tested with 1.24.0)
- [Task](https://taskfile.dev/) installed (`brew install go-task/tap/go-task`)
- AWS CLI configured
- AWS CDK installed (for deployment)

### Local Development

```bash
# List all available tasks
task

# Run all tests
task test

# Run tests with coverage
task test:coverage

# Run integration tests
task test:integration

# Run linter
task lint

# Format code
task fmt

# Full development test cycle
task dev:full-test
```

### Testing the API

```bash
# Register a user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","name":"Test User","password":"password123"}'

# Login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# Access protected route (use token from login)
curl -X GET http://localhost:8080/protected \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Update user profile
curl -X PUT http://localhost:8080/user \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Name"}'

# List users
curl -X GET http://localhost:8080/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Building for Lambda

```bash
# Build Lambda deployment package
task build:lambda

# This creates: .bin/lambda-function.zip
```

### Deploying to AWS

```bash
# Bootstrap CDK (first time only)
task cdk:bootstrap

# Deploy infrastructure and application
task cdk:deploy

# Preview deployment changes
task deploy:diff

# Full deployment workflow (test, build, deploy)
task deploy

# Destroy stack (use with caution!)
task cdk:destroy
```

## Configuration

Set environment variables for configuration:

| Variable         | Description                   | Default              |
|------------------|-------------------------------|----------------------|
| `JWT_SECRET`     | JWT signing secret            | "default-secret-key" |
| `JWT_EXPIRY`     | Token expiry duration         | 24h                  |
| `DYNAMODB_TABLE` | DynamoDB table name           | "users"              |
| `AWS_REGION`     | AWS region                    | "us-east-1"          |
| `ENVIRONMENT`    | "production" or "development" | "development"        |
| `PORT`           | Server port (local only)      | 8080                 |

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./pkg/auth/...
```

### Integration Tests

```bash
# Run all integration tests
task test:integration

# Run all tests including integration
task test:all

# Run specific test suites
task test:handlers    # Handler tests only
task test:auth        # Authentication tests only
task test:database    # Database tests only
task test:models      # Model tests only
```

Test coverage includes:
- ✅ Handler layer tests
- ✅ Service layer tests
- ✅ Database layer tests (mock)
- ✅ Authentication & middleware tests
- ✅ Full user journey integration tests
- ✅ Security & authorization tests

## Key Components

### Config (`pkg/config/`)
Centralized configuration loading from environment variables with typed structs.

### Errors (`internal/errors/` & `pkg/errors/`)
Domain-specific and reusable error definitions with proper error wrapping.

### Validation (`internal/validation/`)
Shared validation logic for user input with clear error messages.

### Authentication (`pkg/auth/`)
JWT token generation and validation with configurable expiry.

### Middleware (`pkg/middleware/`)
JWT authentication middleware for protecting routes.

### Logging (`pkg/logger/`)
Structured logging with slog, including request duration tracking.

## AWS Infrastructure

Deployed resources (via CDK):

- **DynamoDB Table**: `glad-entities` table
- **Lambda Function**: Go 1.x runtime with provided.al2023
- **API Gateway**: REST API with CORS enabled
- **IAM Roles**: Least-privilege access for Lambda

## Development Workflow

1. Write code following the layered architecture
2. Add unit tests for new functionality
3. Format and lint code (`task dev:quick-test`)
4. Run full test suite (`task dev:full-test`)
5. Run integration tests (`task test:integration`)
6. Build Lambda package (`task build:lambda`)
7. Deploy to AWS (`task deploy`)

## Code Quality

- **Linting**: golangci-lint with 10+ linters
- **Test Coverage**: Unit and integration tests
- **CI/CD**: GitHub Actions workflow
- **Security**: JWT authentication, bcrypt password hashing
- **Logging**: Structured logging throughout all layers

## Contributing

This is a learning project. Feel free to explore, fork, and experiment!

## License

This project is for educational purposes.

---

**Status**: Production Ready ✅

All tests passing. Clean architecture. Comprehensive error handling. Security best practices implemented.