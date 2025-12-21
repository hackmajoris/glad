package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AppStackProps struct {
	awscdk.StackProps
}

func NewAppStack(scope constructs.Construct, id string, props *AppStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	ENVIRONMENT := "production"
	awscdk.Tags_Of(stack).Add(jsii.String("Environment"), jsii.String(ENVIRONMENT), nil)

	// Import table from database stack
	tableName := awscdk.Fn_ImportValue(jsii.String("GladTableName"))
	tableArn := awscdk.Fn_ImportValue(jsii.String("GladTableArn"))

	// Create Lambda using Docker image
	myFunc := awslambda.NewDockerImageFunction(stack, jsii.String(id+"-go-func"), &awslambda.DockerImageFunctionProps{
		Code: awslambda.DockerImageCode_FromImageAsset(jsii.String("../../"), &awslambda.AssetImageCodeProps{
			File: jsii.String("Dockerfile.lambda"),
		}),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:   jsii.Number(512),
		Description:  jsii.String("GLAD Lambda function using Docker image"),
		Architecture: awslambda.Architecture_X86_64(),
	})

	myFunc.AddEnvironment(jsii.String("ENVIRONMENT"), jsii.String(ENVIRONMENT), nil)
	myFunc.AddEnvironment(jsii.String("DYNAMODB_TABLE"), tableName, nil)

	// Grant Lambda access to DynamoDB table
	myFunc.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"dynamodb:PutItem",
			"dynamodb:GetItem",
			"dynamodb:UpdateItem",
			"dynamodb:DeleteItem",
			"dynamodb:Query",
			"dynamodb:Scan",
		),
		Resources: jsii.Strings(
			*tableArn,
			*tableArn+"/index/*",
		),
	}))

	api := awsapigateway.NewRestApi(stack, jsii.String(id+"-api-gateway"), &awsapigateway.RestApiProps{
		RestApiName:    jsii.String("glad-api gateway"),
		Description:    jsii.String("GLAD Stack API"),
		Deploy:         jsii.Bool(false),
		CloudWatchRole: jsii.Bool(true),
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins:     jsii.Strings("*"),
			AllowCredentials: jsii.Bool(true),
			AllowHeaders:     jsii.Strings("Content-Type", "Authorization"),
			AllowMethods:     jsii.Strings("GET", "POST", "DELETE", "PUT", "OPTIONS"),
		},
	})

	integration := awsapigateway.NewLambdaIntegration(myFunc, &awsapigateway.LambdaIntegrationOptions{
		Proxy: jsii.Bool(true),
	})

	// Add single wildcard permission for all API Gateway methods
	myFunc.AddPermission(jsii.String("ApiGatewayInvoke"), &awslambda.Permission{
		Principal: awsiam.NewServicePrincipal(jsii.String("apigateway.amazonaws.com"), nil),
		Action:    jsii.String("lambda:InvokeFunction"),
		SourceArn: jsii.String(fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*",
			*stack.Region(),
			*stack.Account(),
			*api.RestApiId())),
	})

	// Define API routes
	registerResource := api.Root().AddResource(jsii.String("register"), nil)
	registerResource.AddMethod(jsii.String("POST"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	loginResource := api.Root().AddResource(jsii.String("login"), nil)
	loginResource.AddMethod(jsii.String("POST"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	protectedResource := api.Root().AddResource(jsii.String("protected"), nil)
	protectedResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	userResource := api.Root().AddResource(jsii.String("user"), nil)
	userResource.AddMethod(jsii.String("PUT"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	usersResource := api.Root().AddResource(jsii.String("users"), nil)
	usersResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	meResource := api.Root().AddResource(jsii.String("me"), nil)
	meResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Skill Management Endpoints
	usersSkillsResource := usersResource.AddResource(jsii.String("{username}"), nil)
	skillsResource := usersSkillsResource.AddResource(jsii.String("skills"), nil)
	skillsResource.AddMethod(jsii.String("POST"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})
	skillsResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	skillResource := skillsResource.AddResource(jsii.String("{skillName}"), nil)
	skillResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})
	skillResource.AddMethod(jsii.String("PUT"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})
	skillResource.AddMethod(jsii.String("DELETE"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Global skill query endpoint
	skillsGlobalResource := api.Root().AddResource(jsii.String("skills"), nil)
	skillNameResource := skillsGlobalResource.AddResource(jsii.String("{skillName}"), nil)
	usersWithSkillResource := skillNameResource.AddResource(jsii.String("users"), nil)
	usersWithSkillResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Master Skills Management Endpoints
	masterSkillsResource := api.Root().AddResource(jsii.String("master-skills"), nil)
	masterSkillsResource.AddMethod(jsii.String("POST"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})
	masterSkillsResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	masterSkillResource := masterSkillsResource.AddResource(jsii.String("{skillID}"), nil)
	masterSkillResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})
	masterSkillResource.AddMethod(jsii.String("PUT"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})
	masterSkillResource.AddMethod(jsii.String("DELETE"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Create deployment
	deployment := awsapigateway.NewDeployment(stack, jsii.String(id+"-api-deployment"), &awsapigateway.DeploymentProps{
		Api:         api,
		Description: jsii.String("Deployment triggered by Lambda changes"),
	})
	deployment.Node().AddDependency(myFunc)

	// Create stage with fixed logical ID
	stage := awsapigateway.NewStage(stack, jsii.String(id+"-api-stage"), &awsapigateway.StageProps{
		Deployment:           deployment,
		StageName:            jsii.String("prod"),
		ThrottlingBurstLimit: jsii.Number(200),
		ThrottlingRateLimit:  jsii.Number(100),
		LoggingLevel:         awsapigateway.MethodLoggingLevel_INFO,
		DataTraceEnabled:     jsii.Bool(true),
		MetricsEnabled:       jsii.Bool(true),
	})

	// Create UsagePlan
	awsapigateway.NewUsagePlan(stack, jsii.String(id+"-api-gateway-usage-plan"), &awsapigateway.UsagePlanProps{
		Name:        jsii.String(id + "-api-gateway-usage-plan"),
		Description: jsii.String("Usage plan with rate limiting"),
		Throttle: &awsapigateway.ThrottleSettings{
			RateLimit:  jsii.Number(100),
			BurstLimit: jsii.Number(200),
		},
		ApiStages: &[]*awsapigateway.UsagePlanPerApiStage{
			{
				Api:   api,
				Stage: stage,
			},
		},
	})

	// Output the API URL
	awscdk.NewCfnOutput(stack, jsii.String("ApiUrl"), &awscdk.CfnOutputProps{
		Value:       jsii.String(fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s", *api.RestApiId(), *stack.Region(), *stage.StageName())),
		Description: jsii.String("API Gateway endpoint URL"),
		ExportName:  jsii.String("GladApiUrl"),
	})

	return stack
}
