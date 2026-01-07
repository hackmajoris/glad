package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AuthStackProps struct {
	awscdk.StackProps
}

func NewAuthStack(scope constructs.Construct, id string, props *AuthStackProps, env string) awscdk.Stack {
	var sprops awscdk.StackProps

	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, &id, &sprops)

	awscdk.Tags_Of(stack).Add(jsii.String("Environment"), jsii.String(env), nil)

	// Import DynamoDB table name from database stack
	tableName := awscdk.Fn_ImportValue(jsii.String("GladTableName-" + env))
	tableArn := awscdk.Fn_ImportValue(jsii.String("GladTableArn-" + env))

	// Create Lambda function for Post Confirmation trigger
	postConfirmationLogGroup := awslogs.NewLogGroup(stack, jsii.String(id+"-post-confirmation-log-group"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/lambda/glad-post-confirmation-" + env),
		Retention:     awslogs.RetentionDays_ONE_WEEK,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	postConfirmationLambda := awslambda.NewFunction(stack, jsii.String(id+"-post-confirmation-lambda"), &awslambda.FunctionProps{
		FunctionName: jsii.String("glad-post-confirmation-" + env),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code: awslambda.Code_FromAsset(jsii.String("../../cmd/cognito-triggers/post-confirmation"), &awss3assets.AssetOptions{
			Bundling: &awscdk.BundlingOptions{
				Image: awslambda.Runtime_PROVIDED_AL2023().BundlingImage(),
				Command: jsii.Strings(
					"bash", "-c",
					"GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOCACHE=/tmp/go-cache GOMODCACHE=/tmp/go-mod go build -ldflags='-s -w' -o /asset-output/bootstrap .",
				),
			},
		}),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(10)),
		MemorySize:   jsii.Number(256),
		Architecture: awslambda.Architecture_X86_64(),
		Environment: &map[string]*string{
			"DYNAMODB_TABLE": tableName,
		},
		LogGroup: postConfirmationLogGroup,
	})

	// Grant DynamoDB write permissions to Post Confirmation Lambda
	postConfirmationLambda.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"dynamodb:PutItem",
			"dynamodb:GetItem",
		),
		Resources: jsii.Strings(*tableArn),
	}))

	// Create Cognito User Pool
	userPool := awscognito.NewUserPool(stack, jsii.String(id+"-user-pool"), &awscognito.UserPoolProps{
		UserPoolName: jsii.String("glad-user-pool-" + env),

		// Sign-in options: Username OR Email
		SignInAliases: &awscognito.SignInAliases{
			Username: jsii.Bool(true),
			Email:    jsii.Bool(true),
		},

		// Self sign-up enabled
		SelfSignUpEnabled: jsii.Bool(true),

		// Password policy: Min 8 chars with complexity requirements
		PasswordPolicy: &awscognito.PasswordPolicy{
			MinLength:            jsii.Number(8),
			RequireUppercase:     jsii.Bool(true),
			RequireLowercase:     jsii.Bool(true),
			RequireDigits:        jsii.Bool(true),
			RequireSymbols:       jsii.Bool(true),
			TempPasswordValidity: awscdk.Duration_Days(jsii.Number(7)),
		},

		// Auto-verify email
		AutoVerify: &awscognito.AutoVerifiedAttrs{
			Email: jsii.Bool(true),
		},

		// Required standard attributes
		StandardAttributes: &awscognito.StandardAttributes{
			Email: &awscognito.StandardAttribute{
				Required: jsii.Bool(true),
				Mutable:  jsii.Bool(true),
			},
		},

		// MFA configuration (optional)
		Mfa: awscognito.Mfa_OPTIONAL,
		MfaSecondFactor: &awscognito.MfaSecondFactor{
			Sms: jsii.Bool(false),
			Otp: jsii.Bool(true), // TOTP via authenticator apps
		},

		// Account recovery
		AccountRecovery: awscognito.AccountRecovery_EMAIL_ONLY,

		// Email settings (using Cognito default for now)
		Email: awscognito.UserPoolEmail_WithCognito(jsii.String("noreply@verificationemail.com")),

		// Advanced security features
		AdvancedSecurityMode: awscognito.AdvancedSecurityMode_ENFORCED,

		// Prevent user existence errors (security best practice)
		SignInCaseSensitive: jsii.Bool(false),

		// Lambda triggers
		LambdaTriggers: &awscognito.UserPoolTriggers{
			// Post Confirmation trigger - creates DynamoDB user profile after email confirmation
			PostConfirmation: postConfirmationLambda,
		},

		// Removal policy for development
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN,
	})

	// Create User Pool Client for Angular application
	userPoolClient := userPool.AddClient(jsii.String(id+"-app-client"), &awscognito.UserPoolClientOptions{
		UserPoolClientName: jsii.String("glad-angular-client-" + env),

		// Auth flows
		AuthFlows: &awscognito.AuthFlow{
			UserPassword: jsii.Bool(true), // USER_PASSWORD_AUTH
			UserSrp:      jsii.Bool(true), // USER_SRP_AUTH (more secure)
		},

		// OAuth 2.0 configuration
		OAuth: &awscognito.OAuthSettings{
			Flows: &awscognito.OAuthFlows{
				AuthorizationCodeGrant: jsii.Bool(true), // OAuth 2.0 + PKCE
			},
			Scopes: &[]awscognito.OAuthScope{
				awscognito.OAuthScope_OPENID(),
				awscognito.OAuthScope_EMAIL(),
				awscognito.OAuthScope_PROFILE(),
			},
			// Callback URLs will be updated after CloudFront distribution is created
			CallbackUrls: jsii.Strings(
				"http://localhost:4200",          // Local development
				"http://localhost:4200/callback", // Local development callback
			),
			LogoutUrls: jsii.Strings(
				"http://localhost:4200",
			),
		},

		// Token validity
		IdTokenValidity:      awscdk.Duration_Hours(jsii.Number(1)), // 1 hour
		AccessTokenValidity:  awscdk.Duration_Hours(jsii.Number(1)), // 1 hour
		RefreshTokenValidity: awscdk.Duration_Days(jsii.Number(30)), // 30 days

		// Prevent user existence errors
		PreventUserExistenceErrors: jsii.Bool(true),

		// Enable refresh token rotation for better security
		EnableTokenRevocation: jsii.Bool(true),
	})

	// Optional: Create User Pool Domain for Hosted UI (if needed later)
	// Commenting out for now as we'll use custom UI in Angular
	/*
		userPool.AddDomain(jsii.String(id+"-user-pool-domain"), &awscognito.UserPoolDomainOptions{
			CognitoDomain: &awscognito.CognitoDomainOptions{
				DomainPrefix: jsii.String("glad-" + env),
			},
		})
	*/

	// CloudFormation outputs
	awscdk.NewCfnOutput(stack, jsii.String("UserPoolId"), &awscdk.CfnOutputProps{
		Value:       userPool.UserPoolId(),
		Description: jsii.String("Cognito User Pool ID"),
		ExportName:  jsii.String("GladUserPoolId-" + env),
	})

	awscdk.NewCfnOutput(stack, jsii.String("UserPoolClientId"), &awscdk.CfnOutputProps{
		Value:       userPoolClient.UserPoolClientId(),
		Description: jsii.String("Cognito User Pool Client ID"),
		ExportName:  jsii.String("GladUserPoolClientId-" + env),
	})

	awscdk.NewCfnOutput(stack, jsii.String("UserPoolArn"), &awscdk.CfnOutputProps{
		Value:       userPool.UserPoolArn(),
		Description: jsii.String("Cognito User Pool ARN"),
		ExportName:  jsii.String("GladUserPoolArn-" + env),
	})

	// Output User Pool region
	awscdk.NewCfnOutput(stack, jsii.String("UserPoolRegion"), &awscdk.CfnOutputProps{
		Value:       stack.Region(),
		Description: jsii.String("AWS Region for Cognito User Pool"),
		ExportName:  jsii.String("GladUserPoolRegion-" + env),
	})

	return stack
}
