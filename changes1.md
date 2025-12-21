# Migrate Lambda Deployment from ZIP to Docker Container

## Summary

Migrated the GLAD Stack Lambda function deployment from ZIP-based packaging to Docker container images. This change provides better dependency management, larger size limits (10GB vs 250MB), and improved local testing capabilities with full Lambda Runtime Interface Emulator support.

## Implementation Details

### New Features

#### 1. Multi-Stage Dockerfile
- **File**: `Dockerfile`
- **Stage 1 (Builder)**: Compiles Go application with static binary optimizations
  - Uses `golang:1.24.0-alpine` base image
  - Builds with `-ldflags="-s -w"` for minimal binary size
  - Produces statically-linked executable
- **Stage 2 (Runtime)**: AWS Lambda runtime environment
  - Uses `public.ecr.aws/lambda/provided:al2023` base image
  - Copies only the compiled binary (minimal attack surface)
  - Sets up Lambda bootstrap handler

#### 2. Enhanced Task Automation
- **File**: `Taskfile.yml`
- **New Tasks**:
  - `task build:docker` - Build Docker image locally for testing
  - `task build:docker:test` - Test Docker image with Lambda Runtime Interface Emulator
- **Updated Tasks**:
  - `task deploy` - Now uses Docker-based deployment (CDK handles build automatically)
  - `task cdk:deploy`, `task cdk:synth`, `task cdk:diff` - Removed ZIP packaging dependencies

#### 3. Automated ECR Integration
- CDK automatically creates and manages ECR repository
- Image versioning based on source code hash (automatic cache invalidation)
- Zero manual Docker registry operations required

### Key Changes

#### Infrastructure (CDK)
- **File**: `deployments/app/cdk.go`
- **Before**: `awslambda.NewFunction()` with ZIP-based `awslambda.AssetCode_FromAsset()`
- **After**: `awslambda.NewDockerImageFunction()` with `awslambda.DockerImageCode_FromImageAsset()`
- **Benefit**: CDK handles entire Docker build → push → deploy pipeline automatically

#### Build Process
- **Before**: Manual `go build` + `zip` commands
- **After**: Dockerfile-based build with CDK orchestration
- **Benefit**: Consistent builds across environments, reproducible deployments

#### Deployment Workflow
CDK now automatically:
1. Builds Docker image from Dockerfile
2. Creates/reuses ECR repository (managed by CDK bootstrap)
3. Tags image with content hash
4. Pushes image to ECR
5. Updates Lambda function to use new image
6. Updates API Gateway integration

## Testing

### Local Testing
```bash
# Build Docker image locally
task build:docker

# Test with Lambda Runtime Interface Emulator
task build:docker:test
```

### Deployment Testing
```bash
# Full deployment (test → build → deploy)
task deploy

# Or deploy directly
task cdk:deploy
```

### Comparison: ZIP vs Docker

| Aspect            | ZIP (Old)             | Docker (New)              |
|-------------------|-----------------------|---------------------------|
| Max Size          | 250MB uncompressed    | 10GB                      |
| Build Process     | Manual go build + zip | Automatic via CDK         |
| Deployment Method | Upload ZIP to Lambda  | Push to ECR, Lambda pulls |
| Dependencies      | Go-only               | Any OS dependencies       |
| Local Testing     | Limited               | Full Lambda emulation     |
| Consistency       | Build environment varies | Same image everywhere  |

## Benefits

1. **Larger Dependency Support**: 10GB limit vs 250MB for ZIP files
2. **Custom Runtime Control**: Full control over OS, libraries, and system dependencies
3. **Development Consistency**: Same Docker image runs locally and in production
4. **Simplified Deployment**: CDK handles entire Docker workflow automatically
5. **Better Testing**: Lambda Runtime Interface Emulator provides accurate local testing
6. **Automatic Versioning**: Image tags based on source code hash ensure proper cache invalidation

## Prerequisites

- Docker must be running (Docker Desktop on macOS)
- AWS CDK bootstrap completed in target account/region
- Sufficient ECR storage quota in AWS account
