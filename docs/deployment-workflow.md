# Deployment Workflow

This document describes the deployment workflow for the GLAD Stack application, focusing on how AWS Cognito configuration is synced to the frontend.

## Overview

The deployment process follows these steps:

1. **Deploy Backend** - Deploy AWS infrastructure (Cognito, DynamoDB, Lambda, API Gateway, etc.)
2. **Generate Amplify Config** - Extract Cognito configuration and generate `amplify_outputs.json`
3. **Build Frontend** - Build Angular application with Amplify configuration
4. **Deploy Frontend** - Upload to S3 and invalidate CloudFront cache

## Amplify Configuration Generation

### What is `amplify_outputs.json`?

The `amplify_outputs.json` file is an Amplify Gen2 compatible configuration file that contains AWS Cognito authentication settings for the frontend. It includes:

- User Pool ID, Client ID, and Region
- Password policy requirements
- MFA configuration
- Username and verification attributes
- Required user attributes

### Generation Tool

The configuration is generated using a Go CLI tool located at `cmd/generate-amplify-config/main.go`.

#### Usage

```bash
# Generate with default settings (production environment)
task generate:amplify-config

# Or run directly with custom options
go run ./cmd/generate-amplify-config/main.go \
  --env production \
  --output site/glad-ui/amplify_outputs.json \
  --region us-east-1
```

#### Command-line Options

- `--env` - Environment name (default: `production`)
- `--output` - Output file path (default: `site/glad-ui/amplify_outputs.json`)
- `--region` - AWS region (defaults to AWS CLI default region)

### How It Works

1. **Fetch CloudFormation Outputs**
   - Connects to AWS CloudFormation
   - Retrieves User Pool ID, Client ID, and Region from the auth stack outputs

2. **Query Cognito Configuration**
   - Uses AWS Cognito IDP API to fetch detailed User Pool configuration
   - Extracts password policy, MFA settings, attributes, etc.

3. **Generate JSON**
   - Transforms the configuration into Amplify Gen2 format
   - Writes to `site/glad-ui/amplify_outputs.json`

### Example Output

```json
{
  "auth": {
    "user_pool_id": "eu-central-1_ABC123",
    "aws_region": "eu-central-1",
    "user_pool_client_id": "abc123def456",
    "mfa_methods": ["TOTP"],
    "standard_required_attributes": ["email"],
    "username_attributes": ["email"],
    "user_verification_types": ["email"],
    "mfa_configuration": "OPTIONAL",
    "password_policy": {
      "min_length": 8,
      "require_lowercase": true,
      "require_numbers": true,
      "require_symbols": true,
      "require_uppercase": true
    }
  },
  "version": "1.0"
}
```

## Deployment Tasks

### Backend Deployment

```bash
task deploy:backend
```

This deploys all CDK stacks:
- `glad-database-stack-production` - DynamoDB table
- `glad-auth-stack-production` - Cognito User Pool and related resources
- `glad-app-stack-production` - Lambda functions and API Gateway
- `glad-frontend-stack-production` - S3 bucket and CloudFront distribution

### Frontend Deployment

```bash
task deploy:frontend
```

This command:
1. Generates `amplify_outputs.json` from deployed backend (dependency)
2. Builds the Angular application with the configuration
3. Uploads build artifacts to S3
4. Invalidates CloudFront cache

### Full Stack Deployment

```bash
task deploy:full
```

This deploys both backend and frontend in sequence, providing the final application URL at the end.

## Important Notes

### Git Ignore

The `amplify_outputs.json` file is **not** committed to git because it contains environment-specific configuration. It should be added to `.gitignore`:

```gitignore
# Amplify configuration (generated from deployment)
site/glad-ui/amplify_outputs.json
```

### Multiple Environments

To support multiple environments (development, staging, production), you can:

1. Deploy separate CDK stacks for each environment
2. Generate separate configuration files:

```bash
# Development
go run ./cmd/generate-amplify-config/main.go \
  --env development \
  --output site/glad-ui/amplify_outputs.dev.json

# Staging
go run ./cmd/generate-amplify-config/main.go \
  --env staging \
  --output site/glad-ui/amplify_outputs.staging.json

# Production (default)
task generate:amplify-config
```

### Local Development

For local development without AWS deployment, you can manually create an `amplify_outputs.json` file with your development User Pool credentials.

## Troubleshooting

### Error: Stack not found

If you see "Stack glad-auth-stack-production not found", ensure you've deployed the backend:

```bash
task deploy:backend
```

### Error: User Pool not found

If the CloudFormation stack exists but User Pool information is missing:

1. Check that the auth stack deployed successfully
2. Verify the stack outputs exist:

```bash
aws cloudformation describe-stacks \
  --stack-name glad-auth-stack-production \
  --query 'Stacks[0].Outputs'
```

### Permission Issues

Ensure your AWS credentials have the following permissions:
- `cloudformation:DescribeStacks`
- `cognito-idp:DescribeUserPool`

## Integration with Frontend

The Angular application should load and use the `amplify_outputs.json` configuration during initialization. This is typically done in the app initialization or authentication service:

```typescript
import amplifyConfig from '../../../amplify_outputs.json';
import { Amplify } from 'aws-amplify';

Amplify.configure(amplifyConfig);
```

## Comparison with sync:config

The project also has a `sync:config` task that generates `environment.prod.ts`. The differences are:

| Feature | `sync:config` | `generate:amplify-config` |
|---------|--------------|---------------------------|
| Output Format | TypeScript | JSON |
| Target | Angular environment | Amplify configuration |
| Includes API URL | Yes | No |
| Includes detailed Cognito config | No | Yes |
| Used by | Custom services | AWS Amplify SDK |

Both can be used together, but for Amplify SDK integration, use `generate:amplify-config`.