package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	ENVIRONMENT := "production"

	getResourceId := func(input string) string {
		return input + "-" + ENVIRONMENT
	}

	// Create database stack first
	NewDatabaseStack(app, getResourceId("glad-database-stack"), &DatabaseStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	}, ENVIRONMENT)

	// Create authentication stack (Cognito User Pool)
	NewAuthStack(app, getResourceId("glad-auth-stack"), &AuthStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	}, ENVIRONMENT)

	// Create application stack (depends on database and auth stacks)
	NewAppStack(app, getResourceId("glad-app-stack"), &AppStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	}, ENVIRONMENT)

	// Create frontend hosting stack (S3 + CloudFront)
	NewFrontendStack(app, getResourceId("glad-frontend-stack"), &FrontendStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	}, ENVIRONMENT)

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
