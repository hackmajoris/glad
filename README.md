# GLAD Stack - Go, Lambda, ApiGateway, DynamoDB

A comprehensive serverless API platform built with Go,
demonstrating modern Cloud-Native architecture using AWS serverless technologies and production-ready API.

## What is GLAD?

**GLAD** stands for:
- **G**o - Modern, efficient programming language with excellent performance and concurrency
- **L**ambda - AWS serverless compute platform for running code without managing servers
- **A**piGateway - AWS managed API gateway service for creating, deploying, and managing REST APIs
- **D**ynamoDB - AWS NoSQL database (Single Table Design) service providing fast and predictable performance with seamless scalability

This project showcases how these four technologies work together to create a production-ready, scalable, and cost-effective serverless API platform
that can handle millions of requests while maintaining low latency and high availability.

[![Go Version](https://img.shields.io/badge/go-1.24.0-blue)]()

## Features

### User Management
- ✅ User registration with validation (username, name, password)
- ✅ User authentication with JWT tokens
- ✅ User profile updates (name, password changes)
- ✅ List all users in the system
- ✅ Bcrypt password hashing for security
- ✅ Get current authenticated user info

### Skills Management
- ✅ **Master Skills Catalog** - Centralized skill definitions
- ✅ **User Skills** - Assign skills to users with proficiency tracking
- ✅ **Proficiency Levels** - Beginner, Intermediate, Advanced, Expert
- ✅ **Years of Experience** - Track experience per skill
- ✅ **Skill Categories** - Organize skills by category (Programming, DevOps, etc.)
- ✅ **Endorsements** - Track skill endorsement counts
- ✅ **Last Used Date** - Track when skill was last used
- ✅ **Skill Notes** - Add custom notes/comments to user skills
- ✅ **Cross-User Queries** - Find all users with a specific skill
- ✅ **Filter by Proficiency** - Query users by skill and proficiency level

### Architecture & Infrastructure
- ✅ **Serverless Architecture** using AWS Lambda + API Gateway
- ✅ **Single Table DynamoDB Design** with Multi-Key GSI pattern
- ✅ **Clean Architecture** with layered design (Handler → Service → Repository)
- ✅ **Repository Pattern** with DynamoDB and Mock implementations
- ✅ **Comprehensive Testing** - unit, integration, and API tests
- ✅ **Structured Logging** using Go's slog package with component tracking
- ✅ **Infrastructure as Code** with AWS CDK (Go)
- ✅ **JWT Authentication** with configurable token expiry
- ✅ **Automatic Mock/Production** repository switching

## Project Structure

```
glad/
├── cmd/
│   └── app/                        # Lambda application
│       ├── main.go                 # Lambda entry point
│       ├── integration_test.go     # Integration tests
│       ├── testdata/               # Test data files
│       └── internal/               # App-specific code
│           ├── database/           # Repository layer (see Database Layer Organization)
│           ├── dto/                # Request/Response DTOs
│           ├── errors/             # App-specific errors
│           ├── handler/            # HTTP handlers (thin layer)
│           ├── models/             # Domain models
│           ├── router/             # Router abstraction
│           ├── service/            # Business logic
│           └── validation/         # Input validation
├── pkg/                            # Shared public packages
│   ├── auth/                       # JWT token service
│   ├── config/                     # Configuration management
│   ├── errors/                     # Core error utilities
│   ├── logger/                     # Structured logging
│   └── middleware/                 # HTTP middleware
├── deployments/
│   └── app/                        # AWS CDK infrastructure
│       ├── cdk.go                  # CDK stack definition
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

1. **Router** - Route matching and middleware chaining for Lambda
2. **Middleware** - JWT validation, logging, CORS
3. **Handler** - HTTP layer (JSON marshaling/unmarshaling)
4. **Service** - Business logic and validation
5. **Repository** - Data access abstraction (interface-based)
6. **Database** - DynamoDB (production) or Mock (development/testing)

### Design Patterns

- **Layered Architecture** - Clear separation of concerns
- **Repository Pattern** - Interface-based data access with factory
- **Dependency Injection** - Constructor injection throughout
- **DTO Pattern** - Separate request/response types from domain models
- **Service Layer** - Business logic isolated from HTTP concerns
- **Factory Pattern** - Auto-selects Mock vs DynamoDB based on environment
- **Single Responsibility** - Each layer has one clear purpose

## Database Layer Organization

The database package follows a scalable file organization pattern designed for growth to 10+ repositories:

```
cmd/app/internal/database/
├── client.go                              # Repository struct definitions
├── constants.go                           # Table names, GSI constants
├── entity_keys.go                         # Entity ID builders and parsers
├── factory.go                             # Repository factory + unified interface
│
├── user_repository.go                     # UserRepository interface
├── user_repository_dynamodb.go            # DynamoDB implementation
├── user_repository_mock.go                # Mock implementation
│
├── user_skill_repository.go               # SkillRepository interface
├── user_skill_repository_dynamodb.go      # DynamoDB implementation
├── user_skill_repository_mock.go          # Mock implementation
│
├── master_skill_repository.go             # MasterSkillRepository interface
├── master_skill_repository_dynamodb.go    # DynamoDB implementation
└── master_skill_repository_mock.go        # Mock implementation
```

**File Naming Pattern**: `{entity}_repository.go`, `{entity}_repository_{implementation}.go`

This pattern ensures:
- Files are grouped by entity when sorted alphabetically
- Clear separation between interface and implementation
- Easy to locate code: interface → DynamoDB impl → Mock impl
- Scales to 10+ repositories without confusion

**Repository Pattern**: Each entity has:
1. An interface defining operations (e.g., `UserRepository`)
2. DynamoDB implementation using single-table design
3. Mock implementation for local development and testing

The unified `Repository` interface composes all entity repositories, allowing both `DynamoDBRepository` and `MockRepository` to implement the same interface.

**Auto-Selection Logic**:
- If `AWS_LAMBDA_FUNCTION_NAME` environment variable exists → DynamoDB
- If `ENVIRONMENT=production` → DynamoDB
- If `ENVIRONMENT=development` or `DB_MOCK=true` → Mock
- Default: DynamoDB

## Data Model - Single Table Design with Multi-Key GSI

The Single Table Design is modeled using DynamoDB's Multi-Key (composite keys) for GSI feature.
Read more: https://aws.amazon.com/blogs/database/multi-key-support-for-global-secondary-index-in-amazon-dynamodb/

### Table: `glad-entities`
- **Partition Key**: `entity_id` (STRING)

### Entity ID Format (using `#` delimiter):
- **Users**: `USER#<username>` (e.g., `USER#john`)
- **User Skills**: `USERSKILL#<username>#<skill_id>` (e.g., `USERSKILL#john#python`)
- **Master Skills**: `SKILL#<skill_id>` (e.g., `SKILL#python`)

### Global Secondary Indexes (5 GSIs):

1. **SkillsByLevel** - Query users by skill and proficiency level
   - PK: `SkillName`, SK: `ProficiencyLevel`
2. **ByUser** - Get all entities for a user
   - PK: `Username`, SK: `EntityType`
3. **SkillsByCategory** - Find skills by category
   - PK: `EntityType`, SK: `Category`
4. **ByEntityType** - Query all entities of a type
   - PK: `EntityType`, SK: `SkillName`
5. **BySkillID** - Find all users with a specific skill
   - PK: `skill_id`, SK: `Username`

## API Endpoints

### Authentication (Public)
| Method | Path       | Auth | Description                  |
|--------|------------|------|------------------------------|
| POST   | /register  | No   | User registration            |
| POST   | /login     | No   | Authentication (returns JWT) |

### User Management (Protected - JWT Required)
| Method | Path       | Auth | Description                  |
|--------|------------|------|------------------------------|
| GET    | /me        | JWT  | Get current user info        |
| GET    | /users     | JWT  | List all users               |
| PUT    | /user      | JWT  | Update user profile          |
| GET    | /protected | JWT  | Protected resource demo      |

### User Skills (Protected - JWT Required)
| Method | Path                               | Auth | Description              |
|--------|------------------------------------|------|--------------------------|
| POST   | /users/{username}/skills           | JWT  | Add skill to user        |
| GET    | /users/{username}/skills           | JWT  | List all skills for user |
| GET    | /users/{username}/skills/{skillID} | JWT  | Get specific user skill  |
| PUT    | /users/{username}/skills/{skillID} | JWT  | Update user skill        |
| DELETE | /users/{username}/skills/{skillID} | JWT  | Delete user skill        |

### Master Skills (Protected - JWT Required)
| Method | Path                     | Auth | Description               |
|--------|--------------------------|------|---------------------------|
| POST   | /master-skills           | JWT  | Create master skill       |
| GET    | /master-skills           | JWT  | List all master skills    |
| GET    | /master-skills/{skillID} | JWT  | Get specific master skill |
| PUT    | /master-skills/{skillID} | JWT  | Update master skill       |
| DELETE | /master-skills/{skillID} | JWT  | Delete master skill       |

### Cross-User Skill Queries (Protected - JWT Required)
| Method | Path                                   | Auth | Description                    |
|--------|----------------------------------------|------|--------------------------------|
| GET    | /skills/{skillName}/users              | JWT  | Find all users with skill      |
| GET    | /skills/{skillName}/users?level=Expert | JWT  | Find users with skill at level |

## Getting Started

### Prerequisites

- Go 1.21+ (tested with 1.24.0)
- [Task](https://taskfile.dev/) installed (`brew install go-task/tap/go-task`)
- AWS CLI configured with credentials
- AWS CDK installed (for deployment): `npm install -g aws-cdk`

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

# Quick test cycle (format + test)
task dev:quick-test

# Full development test cycle (format + lint + test + build)
task dev:full-test
```

### Testing the API

#### User Registration & Authentication
```bash
# Register a user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","name":"Test User","password":"password123"}'

# Login (returns JWT token)
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# Get current user info
curl -X GET http://localhost:8080/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Update user profile
curl -X PUT http://localhost:8080/user \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Name"}'

# List all users
curl -X GET http://localhost:8080/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### Master Skills Management
```bash
# Create a master skill
curl -X POST http://localhost:8080/master-skills \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "skill_id":"python",
    "skill_name":"Python",
    "category":"Programming"
  }'

# List all master skills
curl -X GET http://localhost:8080/master-skills \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

#### User Skills Management
```bash
# Add skill to user
curl -X POST http://localhost:8080/users/testuser/skills \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "skill_id":"python",
    "skill_name":"Python",
    "category":"Programming",
    "proficiency_level":"Intermediate",
    "years_of_experience":3
  }'

# Get user's skills
curl -X GET http://localhost:8080/users/testuser/skills \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Find all users with Python skill
curl -X GET http://localhost:8080/skills/Python/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Find Expert Python developers
curl -X GET "http://localhost:8080/skills/Python/users?level=Expert" \
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
# Bootstrap CDK (first time only, per account/region)
task cdk:bootstrap

# Preview deployment changes
task cdk:diff

# Deploy infrastructure and application
task cdk:deploy

# Full deployment workflow (test → build → deploy)
task deploy

# Destroy stack (use with caution!)
task cdk:destroy
```

## Configuration

Set environment variables for configuration:

| Variable                   | Description                   | Default              |
|----------------------------|-------------------------------|----------------------|
| `JWT_SECRET`               | JWT signing secret            | "default-secret-key" |
| `JWT_EXPIRY`               | Token expiry duration         | 24h                  |
| `JWT_SIGNING_ALG`          | JWT signing algorithm         | "HS256"              |
| `DYNAMODB_TABLE`           | DynamoDB table name           | "users"              |
| `AWS_REGION`               | AWS region for DynamoDB       | "us-east-1"          |
| `ENVIRONMENT`              | "production" or "development" | "development"        |
| `PORT`                     | Server port (local only)      | 8080                 |
| `DB_MOCK`                  | Force mock DB usage           | (not set)            |
| `AWS_LAMBDA_FUNCTION_NAME` | Auto-detected in Lambda       | (auto)               |

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./pkg/auth/...

# Run with race detection
go test -race ./...
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

### Test Coverage

```bash
# Generate coverage report
task test:coverage

# View coverage in browser
go tool cover -html=coverage.out
```

Test coverage includes:
- ✅ Handler layer tests (user, skill, master skill)
- ✅ Service layer tests
- ✅ Database layer tests (Mock repository)
- ✅ Authentication & middleware tests
- ✅ Full user journey integration tests
- ✅ Security & authorization tests
- ✅ Domain model validation tests

## Key Components

### Config (`pkg/config/`)
Centralized configuration loading from environment variables with typed structs and defaults.

### Errors (`cmd/app/internal/errors/` & `pkg/errors/`)
- Domain-specific error definitions
- Reusable error utilities
- HTTP status code mapping
- Proper error wrapping with context

### Validation (`cmd/app/internal/validation/`)
- Username validation (3-50 chars, alphanumeric + underscore)
- Password validation (min 6 chars)
- Name validation (non-empty)
- Skill ID validation (lowercase alphanumeric + dashes)
- Proficiency level enum validation
- Years of experience validation (non-negative)

### Authentication (`pkg/auth/`)
- JWT token generation with configurable expiry
- Token validation and claims extraction
- HS256 signing algorithm
- Username embedded in token claims

### Middleware (`pkg/middleware/`)
- JWT authentication middleware
- Bearer token extraction from Authorization header
- Route protection
- Error handling in auth flow

### Logging (`pkg/logger/`)
- Structured logging with Go's slog package
- Component-based logging (e.g., "database", "handler")
- Operation tracking
- Duration tracking for performance monitoring
- Levels: Info (✅), Debug (🔍), Error (❌), Warn (⚠️)
- JSON format for production, text for development

## AWS Infrastructure

Deployed resources (via AWS CDK in Go):

### DynamoDB Table
- **Name**: `glad-entities`
- **Single table** with 5 Global Secondary Indexes
- **DynamoDB Streams**: Enabled (NEW_AND_OLD_IMAGES)
- **Capacity**: On-demand billing mode
- **Point-in-time recovery**: Disabled (dev-friendly)
- **Removal policy**: DESTROY (dev-friendly)

### Lambda Function
- **Runtime**: provided.al2023 (custom Go runtime)
- **Handler**: bootstrap binary
- **Architecture**: AMD64
- **Timeout**: 30 seconds
- **Environment**: ENVIRONMENT=production
- **Permissions**: DynamoDB read/write on `glad-entities`

### API Gateway
- **Type**: REST API
- **Name**: glad-api-gateway
- **CORS**: Enabled (* origins)
- **Throttling**: 100 RPS, 200 burst
- **Deployment**: Production stage
- **Usage Plan**: Attached with rate limiting

### IAM Roles
- Lambda execution role with DynamoDB permissions
- Least-privilege access pattern

## Development Workflow

1. **Write code** following the layered architecture
2. **Add unit tests** for new functionality
3. **Format and lint** code (`task dev:quick-test`)
4. **Run full test suite** (`task dev:full-test`)
5. **Run integration tests** (`task test:integration`)
6. **Build Lambda package** (`task build:lambda`)
7. **Preview changes** (`task cdk:diff`)
8. **Deploy to AWS** (`task deploy`)

## Code Quality

- **Linting**: golangci-lint via `task lint`
- **Test Coverage**: Unit and integration tests across all layers
- **CI/CD**: GitHub Actions workflow (see `.github/workflows/`)
- **Security**:
  - JWT authentication with Bearer tokens
  - Bcrypt password hashing (cost: 10)
  - Input validation on all endpoints
  - Proper error handling without leaking sensitive data
- **Logging**: Structured logging throughout all layers
- **Error Handling**: Domain-specific errors with HTTP mapping

## Dependencies

### Go Packages
- `github.com/aws/aws-lambda-go` - Lambda runtime
- `github.com/aws/aws-sdk-go` - DynamoDB client
- `github.com/golang-jwt/jwt/v5` - JWT token handling
- `golang.org/x/crypto` - Bcrypt password hashing

### AWS Services
- AWS Lambda (serverless compute)
- API Gateway (REST API management)
- DynamoDB (NoSQL database with single-table design)
- CloudFormation (infrastructure via CDK)
- IAM (permissions management)

## Contributing

This is a learning project demonstrating serverless Go architecture. Feel free to explore, fork, and experiment!

## License

This project is for educational purposes.

---