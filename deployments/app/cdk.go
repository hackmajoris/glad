package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CdkStackProps struct {
	awscdk.StackProps
}

func createEntitiesTable(stack awscdk.Stack, id *string, environment string) awsdynamodb.TableV2 {
	entitiesTable := awsdynamodb.NewTableV2(stack, id, &awsdynamodb.TablePropsV2{
		TableName: jsii.String("glad-entities"),
		// Partition Key: EntityType
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("EntityType"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("entity_id"),
			Type: awsdynamodb.AttributeType_STRING,
		},

		GlobalSecondaryIndexes: &[]*awsdynamodb.GlobalSecondaryIndexPropsV2{
			// GSI for flexible category/skill/proficiency queries
			// Single PK: Category (allows broad queries)
			// Composite SK: SkillName + ProficiencyLevel + YearsOfExperience + Username
			// This design provides maximum query flexibility:
			//   - Query by Category alone
			//   - Query by Category + SkillName
			//   - Query by Category + SkillName + ProficiencyLevel
			//   - Query by Category + SkillName + ProficiencyLevel + YearsOfExperience (with range)
			{
				IndexName: jsii.String("BySkill"),
				PartitionKey: &awsdynamodb.Attribute{
					Name: jsii.String("Category"),
					Type: awsdynamodb.AttributeType_STRING,
				},
				SortKeys: &[]*awsdynamodb.Attribute{
					{
						Name: jsii.String("SkillName"),
						Type: awsdynamodb.AttributeType_STRING,
					},
					{
						Name: jsii.String("ProficiencyLevel"),
						Type: awsdynamodb.AttributeType_STRING,
					},
					{
						Name: jsii.String("YearsOfExperience"),
						Type: awsdynamodb.AttributeType_NUMBER,
					},
					{
						Name: jsii.String("Username"),
						Type: awsdynamodb.AttributeType_STRING,
					},
				},
			},
		},

		PointInTimeRecovery: jsii.Bool(false),
		DynamoStream:        awsdynamodb.StreamViewType_NEW_AND_OLD_IMAGES,
		RemovalPolicy:       awscdk.RemovalPolicy_DESTROY,

		Tags: &[]*awscdk.CfnTag{
			{
				Key:   jsii.String("Purpose"),
				Value: jsii.String("Single-Table-Design"),
			},
			{
				Key:   jsii.String("DataModel"),
				Value: jsii.String("Multi-Entity"),
			},
		},
	})

	return entitiesTable
}

func NewCdkStack(scope constructs.Construct, id string, props *CdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	ENVIRONMENT := "production" // todo: will be parametrised

	// Add environment tag
	awscdk.Tags_Of(stack).Add(jsii.String("Environment"), jsii.String(ENVIRONMENT), nil)

	// Create Lambda using Docker image
	myFunc := awslambda.NewDockerImageFunction(stack, jsii.String(id+"-go-func"), &awslambda.DockerImageFunctionProps{
		Code: awslambda.DockerImageCode_FromImageAsset(jsii.String("../../"), &awslambda.AssetImageCodeProps{
			File: jsii.String("Dockerfile"),
		}),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:   jsii.Number(512),
		Description:  jsii.String("GLAD Lambda function using Docker image"),
		Architecture: awslambda.Architecture_X86_64(),
	})

	myFunc.AddEnvironment(jsii.String("ENVIRONMENT"), jsii.String(ENVIRONMENT), nil)

	////  Create table | Grant Lambda read/write access to DynamoDB table
	entitiesTable := createEntitiesTable(stack, jsii.String(id+"-entities-table-"+ENVIRONMENT), ENVIRONMENT)
	// Grant access to table and all GSIs with wildcard to avoid policy size issues
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
			*entitiesTable.TableArn(),
			*entitiesTable.TableArn()+"/index/*",
		),
	}))

	api := awsapigateway.NewRestApi(stack, jsii.String(id+"-api-gateway"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("glad-api gateway"),
		Description: jsii.String("GLAD Stack API"),
		// Don't use DeployOptions - we manage deployment explicitly below
		Deploy:         jsii.Bool(false), // Disable automatic deployment
		CloudWatchRole: jsii.Bool(true),  // Auto-create IAM role for CloudWatch Logs
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins:     jsii.Strings("*"),
			AllowCredentials: jsii.Bool(true),
			AllowHeaders:     jsii.Strings("Content-Type", "Authorization"),
			AllowMethods:     jsii.Strings("GET", "POST", "DELETE", "PUT", "OPTIONS"),
		},
	})

	// Create Lambda integration with explicit proxy configuration
	// AWS_PROXY mode passes the entire request to Lambda and expects Lambda to return proper API Gateway response
	integration := awsapigateway.NewLambdaIntegration(myFunc, &awsapigateway.LambdaIntegrationOptions{
		Proxy: jsii.Bool(true), // Explicitly enable AWS_PROXY mode
	})

	// Add single wildcard permission for all API Gateway methods to avoid policy size limit
	myFunc.AddPermission(jsii.String("ApiGatewayInvoke"), &awslambda.Permission{
		Principal: awsiam.NewServicePrincipal(jsii.String("apigateway.amazonaws.com"), nil),
		Action:    jsii.String("lambda:InvokeFunction"),
		SourceArn: jsii.String(fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/*/*",
			*stack.Region(),
			*stack.Account(),
			*api.RestApiId())),
	})

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

	// Add missing /users GET endpoint
	usersResource := api.Root().AddResource(jsii.String("users"), nil)
	usersResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Add /me GET endpoint for current user
	meResource := api.Root().AddResource(jsii.String("me"), nil)
	meResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Skill Management Endpoints
	// Pattern: /users/{username}/skills
	usersSkillsResource := usersResource.AddResource(jsii.String("{username}"), nil)
	skillsResource := usersSkillsResource.AddResource(jsii.String("skills"), nil)

	// POST /users/{username}/skills - Add a skill
	skillsResource.AddMethod(jsii.String("POST"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// GET /users/{username}/skills - List all skills for user
	skillsResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Specific skill endpoints
	// Pattern: /users/{username}/skills/{skillName}
	skillResource := skillsResource.AddResource(jsii.String("{skillName}"), nil)

	// GET /users/{username}/skills/{skillName} - Get specific skill
	skillResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// PUT /users/{username}/skills/{skillName} - Update skill
	skillResource.AddMethod(jsii.String("PUT"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// DELETE /users/{username}/skills/{skillName} - Delete skill
	skillResource.AddMethod(jsii.String("DELETE"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Global skill query endpoint
	// GET /skills/{skillName}/users - Find all users with a skill
	skillsGlobalResource := api.Root().AddResource(jsii.String("skills"), nil)
	skillNameResource := skillsGlobalResource.AddResource(jsii.String("{skillName}"), nil)
	usersWithSkillResource := skillNameResource.AddResource(jsii.String("users"), nil)
	usersWithSkillResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Master Skills Management Endpoints
	// Pattern: /master-skills
	masterSkillsResource := api.Root().AddResource(jsii.String("master-skills"), nil)

	// POST /master-skills - Create a master skill
	masterSkillsResource.AddMethod(jsii.String("POST"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// GET /master-skills - List all master skills
	masterSkillsResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Specific master skill endpoints
	// Pattern: /master-skills/{skillID}
	masterSkillResource := masterSkillsResource.AddResource(jsii.String("{skillID}"), nil)

	// GET /master-skills/{skillID} - Get specific master skill
	masterSkillResource.AddMethod(jsii.String("GET"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// PUT /master-skills/{skillID} - Update master skill
	masterSkillResource.AddMethod(jsii.String("PUT"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// DELETE /master-skills/{skillID} - Delete master skill
	masterSkillResource.AddMethod(jsii.String("DELETE"), integration, &awsapigateway.MethodOptions{
		AuthorizationType: awsapigateway.AuthorizationType_NONE,
	})

	// Force API Gateway to create new deployment when Lambda changes
	// This prevents deployment drift issues when switching between ZIP and Docker images
	deployment := awsapigateway.NewDeployment(stack, jsii.String(id+"-api-deployment"), &awsapigateway.DeploymentProps{
		Api:         api,
		Description: jsii.String("Deployment triggered by Lambda changes"),
	})

	// Add dependency on Lambda function to trigger redeployment when Lambda changes
	deployment.Node().AddDependency(myFunc)

	// Update stage to use the explicit deployment
	// Use fixed logical ID for stable updates
	stage := awsapigateway.NewStage(stack, jsii.String(id+"-api-stage"), &awsapigateway.StageProps{
		Deployment:           deployment,
		StageName:            jsii.String("prod"),
		ThrottlingBurstLimit: jsii.Number(200),
		ThrottlingRateLimit:  jsii.Number(100),
		LoggingLevel:         awsapigateway.MethodLoggingLevel_INFO,
		DataTraceEnabled:     jsii.Bool(true),
		MetricsEnabled:       jsii.Bool(true),
	})

	// Create UsagePlan AFTER all methods are defined
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
				Stage: stage, // Use explicit stage instead of api.DeploymentStage()
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

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewCdkStack(app, "glad-stack", &CdkStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
