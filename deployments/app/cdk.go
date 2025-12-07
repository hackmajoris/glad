package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CdkStackProps struct {
	awscdk.StackProps
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

	// The code that defines your stack goes here

	// example resource
	// queue := awssqs.NewQueue(stack, jsii.String("CdkQueue"), &awssqs.QueueProps{
	// 	VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
	// })

	// Create DynamoDB Single Table
	// This table uses single-table design pattern to store multiple entity types
	// Entities: User, UserSkill (and future: Project, Settings, etc.)
	// Key structure:
	//   - User:      PK=USER#<username>, SK=PROFILE
	//   - UserSkill: PK=USER#<username>, SK=SKILL#<skill_name>

	entitiesTable := awsdynamodb.NewTableV2(stack, jsii.String(id+"-entities-table"), &awsdynamodb.TablePropsV2{
		TableName: jsii.String("glad-entities"),

		// Partition Key: PK (stores entity identifier)
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("PK"),
			Type: awsdynamodb.AttributeType_STRING,
		},

		// Sort Key: SK (stores entity type and sub-identifier)
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("SK"),
			Type: awsdynamodb.AttributeType_STRING,
		},

		// GSI1: For cross-entity queries (e.g., find all users with a skill)
		GlobalSecondaryIndexes: &[]*awsdynamodb.GlobalSecondaryIndexPropsV2{
			{
				IndexName: jsii.String("GSI1"),
				PartitionKey: &awsdynamodb.Attribute{
					Name: jsii.String("GSI1PK"),
					Type: awsdynamodb.AttributeType_STRING,
				},
				SortKey: &awsdynamodb.Attribute{
					Name: jsii.String("GSI1SK"),
					Type: awsdynamodb.AttributeType_STRING,
				},
				// INCLUDE projection for cost optimization
				// Only includes essential attributes needed for queries
				ProjectionType: awsdynamodb.ProjectionType_INCLUDE,
				NonKeyAttributes: jsii.Strings(
					"EntityType",
					"Username",
					"SkillName",
					"ProficiencyLevel",
					"Name",
				),
			},
		},

		// Enable point-in-time recovery for data protection
		PointInTimeRecovery: jsii.Bool(true),

		// Enable DynamoDB Streams for event-driven architecture
		DynamoStream: awsdynamodb.StreamViewType_NEW_AND_OLD_IMAGES,

		// Remove table on stack deletion (for dev/testing)
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,

		// Additional tags
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

	// Create Lambda
	myFunc := awslambda.NewFunction(stack, jsii.String(id+"-go-func"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Code:    awslambda.AssetCode_FromAsset(jsii.String("../../.bin/lambda-function.zip"), nil),
		Handler: jsii.String("main"),
	})

	myFunc.AddEnvironment(jsii.String("environment"), jsii.String(ENVIRONMENT), nil)

	// Grant Lambda read/write access to DynamoDB table
	entitiesTable.GrantReadWriteData(myFunc)

	api := awsapigateway.NewRestApi(stack, jsii.String(id+"-api-gateway"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("glad-api gateway"),
		Description: jsii.String("GLAD Stack API"),
		DeployOptions: &awsapigateway.StageOptions{
			StageName:            jsii.String("prod"),
			ThrottlingBurstLimit: jsii.Number(200),
			ThrottlingRateLimit:  jsii.Number(100),
		},
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins:     jsii.Strings("*"),
			AllowCredentials: jsii.Bool(true),
			AllowHeaders:     jsii.Strings("Content-Type", "Authorization"),
			AllowMethods:     jsii.Strings("GET", "POST", "DELETE", "PUT", "OPTIONS"),
		},
	})

	// Create Lambda integration
	integration := awsapigateway.NewLambdaIntegration(myFunc, nil)

	registerResource := api.Root().AddResource(jsii.String("register"), nil)
	registerResource.AddMethod(jsii.String("POST"), integration, nil)

	loginResource := api.Root().AddResource(jsii.String("login"), nil)
	loginResource.AddMethod(jsii.String("POST"), integration, nil)

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
				Stage: api.DeploymentStage(),
			},
		},
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
