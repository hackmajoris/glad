package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// AmplifyOutputs represents the Amplify Gen2 configuration structure
type AmplifyOutputs struct {
	Auth    AuthConfig   `json:"auth"`
	Custom  CustomConfig `json:"custom,omitempty"`
	Version string       `json:"version"`
}

// AuthConfig represents the authentication configuration
type AuthConfig struct {
	UserPoolID            string         `json:"user_pool_id"`
	AWSRegion             string         `json:"aws_region"`
	UserPoolClientID      string         `json:"user_pool_client_id"`
	MFAMethods            []string       `json:"mfa_methods"`
	StandardRequiredAttrs []string       `json:"standard_required_attributes"`
	UsernameAttributes    []string       `json:"username_attributes"`
	UserVerificationTypes []string       `json:"user_verification_types"`
	MFAConfiguration      string         `json:"mfa_configuration"`
	PasswordPolicy        PasswordPolicy `json:"password_policy"`
}

// CustomConfig represents custom backend configuration
type CustomConfig struct {
	API APIConfig `json:"api"`
}

// APIConfig represents the API Gateway configuration
type APIConfig struct {
	Endpoint string `json:"endpoint"`
	Region   string `json:"region"`
}

// PasswordPolicy represents the password requirements
type PasswordPolicy struct {
	MinLength        int32 `json:"min_length"`
	RequireLowercase bool  `json:"require_lowercase"`
	RequireNumbers   bool  `json:"require_numbers"`
	RequireSymbols   bool  `json:"require_symbols"`
	RequireUppercase bool  `json:"require_uppercase"`
}

func main() {
	var (
		environment = flag.String("env", "production", "Environment name (e.g., production, staging)")
		outputPath  = flag.String("output", "site/glad-ui/amplify_outputs.json", "Output file path")
		region      = flag.String("region", "", "AWS region (defaults to default region)")
	)

	flag.Parse()

	fmt.Println("🔄 Generating amplify_outputs.json from CDK deployment...")
	fmt.Printf("   Environment: %s\n", *environment)
	fmt.Println()

	ctx := context.Background()

	// Load AWS SDK configuration
	var opts []func(*config.LoadOptions) error
	if *region != "" {
		opts = append(opts, config.WithRegion(*region))
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error loading AWS configuration: %v\n", err)
		os.Exit(1)
	}

	if awsConfig.Region == "" {
		awsConfig.Region = "us-east-1"
	}

	fmt.Printf("   Region: %s\n\n", awsConfig.Region)

	// Stack names
	authStackName := fmt.Sprintf("glad-auth-stack-%s", *environment)
	appStackName := fmt.Sprintf("glad-app-stack-%s", *environment)

	// Initialize CloudFormation client
	cfnClient := cloudformation.NewFromConfig(awsConfig)

	// Fetch CloudFormation outputs
	fmt.Println("📋 Fetching CloudFormation outputs...")

	userPoolID, err := getStackOutput(ctx, cfnClient, authStackName, "UserPoolId")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "   Please deploy the auth stack first: cdk deploy %s\n", authStackName)
		os.Exit(1)
	}

	userPoolClientID, err := getStackOutput(ctx, cfnClient, authStackName, "UserPoolClientId")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}

	userPoolRegion, err := getStackOutput(ctx, cfnClient, authStackName, "UserPoolRegion")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}

	// Fetch API endpoint from app stack
	apiUrl, err := getStackOutput(ctx, cfnClient, appStackName, "ApiUrl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Could not fetch API endpoint: %v\n", err)
		fmt.Fprintf(os.Stderr, "   API configuration will be omitted. Deploy the app stack: cdk deploy %s\n", appStackName)
		apiUrl = "" // Continue without API config
	}

	fmt.Println("✅ Fetched CloudFormation outputs:")
	fmt.Printf("   User Pool ID: %s\n", userPoolID)
	fmt.Printf("   User Pool Client ID: %s\n", userPoolClientID)
	fmt.Printf("   User Pool Region: %s\n", userPoolRegion)
	if apiUrl != "" {
		fmt.Printf("   API Endpoint: %s\n", apiUrl)
	}
	fmt.Println()

	// Initialize Cognito client for the User Pool's region
	cognitoConfig := awsConfig.Copy()
	cognitoConfig.Region = userPoolRegion
	cognitoClient := cognitoidentityprovider.NewFromConfig(cognitoConfig)

	// Fetch User Pool configuration
	fmt.Println("📥 Fetching Cognito User Pool configuration...")

	userPoolConfig, err := getUserPoolConfig(ctx, cognitoClient, userPoolID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error fetching User Pool configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Fetched User Pool configuration")
	fmt.Println()

	// Generate amplify_outputs.json
	fmt.Println("📝 Generating amplify_outputs.json...")

	amplifyConfig := generateAmplifyConfig(userPoolID, userPoolClientID, userPoolRegion, userPoolConfig, apiUrl, awsConfig.Region)

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(*outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	file, err := os.Create(*outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(amplifyConfig); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error writing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ amplify_outputs.json generated successfully: %s\n", *outputPath)
	fmt.Println()
	fmt.Println("📋 Configuration summary:")

	// Pretty print the configuration
	prettyJSON, _ := json.MarshalIndent(amplifyConfig, "", "  ")
	fmt.Println(string(prettyJSON))

	fmt.Println()
	fmt.Println("🎉 Configuration generation complete!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Build the Angular app: cd site/glad-ui && npm run build")
	fmt.Println("  2. Deploy to S3: task frontend:deploy")
}

func getStackOutput(ctx context.Context, client *cloudformation.Client, stackName, outputKey string) (string, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: &stackName,
	}

	result, err := client.DescribeStacks(ctx, input)
	if err != nil {
		return "", fmt.Errorf("stack %s not found: %w", stackName, err)
	}

	if len(result.Stacks) == 0 {
		return "", fmt.Errorf("stack %s not found", stackName)
	}

	for _, output := range result.Stacks[0].Outputs {
		if output.OutputKey != nil && *output.OutputKey == outputKey {
			if output.OutputValue != nil {
				return *output.OutputValue, nil
			}
		}
	}

	return "", fmt.Errorf("output %s not found in stack %s", outputKey, stackName)
}

func getUserPoolConfig(ctx context.Context, client *cognitoidentityprovider.Client, userPoolID string) (*types.UserPoolType, error) {
	input := &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: &userPoolID,
	}

	result, err := client.DescribeUserPool(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.UserPool, nil
}

func generateAmplifyConfig(userPoolID, userPoolClientID, region string, userPoolConfig *types.UserPoolType, apiUrl, apiRegion string) *AmplifyOutputs {
	// Extract username attributes
	usernameAttrs := make([]string, 0)
	if userPoolConfig.UsernameAttributes != nil {
		for _, attr := range userPoolConfig.UsernameAttributes {
			usernameAttrs = append(usernameAttrs, string(attr))
		}
	}

	// Extract auto-verified attributes
	autoVerifiedAttrs := make([]string, 0)
	if userPoolConfig.AutoVerifiedAttributes != nil {
		for _, attr := range userPoolConfig.AutoVerifiedAttributes {
			autoVerifiedAttrs = append(autoVerifiedAttrs, string(attr))
		}
	}

	// Extract required standard attributes
	requiredAttrs := make([]string, 0)
	if userPoolConfig.SchemaAttributes != nil {
		for _, attr := range userPoolConfig.SchemaAttributes {
			if attr.Required != nil && *attr.Required {
				if attr.Name != nil && !isCustomAttribute(*attr.Name) {
					requiredAttrs = append(requiredAttrs, *attr.Name)
				}
			}
		}
	}

	// Extract MFA configuration
	mfaConfig := "NONE"
	if userPoolConfig.MfaConfiguration != "" {
		mfaConfig = string(userPoolConfig.MfaConfiguration)
	}

	// Extract MFA methods
	mfaMethods := make([]string, 0)
	if userPoolConfig.MfaConfiguration == types.UserPoolMfaTypeOptional ||
		userPoolConfig.MfaConfiguration == types.UserPoolMfaTypeOn {
		// Check if SMS MFA is configured
		if userPoolConfig.SmsConfiguration != nil {
			mfaMethods = append(mfaMethods, "SMS")
		}
		// Check if Software Token (TOTP) MFA is enabled
		if userPoolConfig.UserPoolAddOns != nil && userPoolConfig.UserPoolAddOns.AdvancedSecurityMode != "" {
			mfaMethods = append(mfaMethods, "TOTP")
		}
	}

	// Extract password policy
	passwordPolicy := PasswordPolicy{
		MinLength:        8,
		RequireLowercase: false,
		RequireNumbers:   false,
		RequireSymbols:   false,
		RequireUppercase: false,
	}

	if userPoolConfig.Policies != nil && userPoolConfig.Policies.PasswordPolicy != nil {
		policy := userPoolConfig.Policies.PasswordPolicy
		if policy.MinimumLength != nil {
			passwordPolicy.MinLength = *policy.MinimumLength
		}
		passwordPolicy.RequireLowercase = policy.RequireLowercase
		passwordPolicy.RequireNumbers = policy.RequireNumbers
		passwordPolicy.RequireSymbols = policy.RequireSymbols
		passwordPolicy.RequireUppercase = policy.RequireUppercase
	}

	config := &AmplifyOutputs{
		Auth: AuthConfig{
			UserPoolID:            userPoolID,
			AWSRegion:             region,
			UserPoolClientID:      userPoolClientID,
			MFAMethods:            mfaMethods,
			StandardRequiredAttrs: requiredAttrs,
			UsernameAttributes:    usernameAttrs,
			UserVerificationTypes: autoVerifiedAttrs,
			MFAConfiguration:      mfaConfig,
			PasswordPolicy:        passwordPolicy,
		},
		Version: "1.0",
	}

	// Add API configuration if available
	if apiUrl != "" {
		config.Custom = CustomConfig{
			API: APIConfig{
				Endpoint: apiUrl,
				Region:   apiRegion,
			},
		}
	}

	return config
}

func isCustomAttribute(name string) bool {
	return len(name) >= 7 && name[:7] == "custom:"
}
