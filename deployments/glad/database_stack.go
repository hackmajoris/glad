package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DatabaseStackProps struct {
	awscdk.StackProps
}

func NewDatabaseStack(scope constructs.Construct, id string, props *DatabaseStackProps, env string) awscdk.Stack {
	var sprops awscdk.StackProps

	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	awscdk.Tags_Of(stack).Add(jsii.String("Environment"), jsii.String(env), nil)

	// Create DynamoDB table
	entitiesTable := awsdynamodb.NewTableV2(stack, jsii.String(id+"-entities-table"), &awsdynamodb.TablePropsV2{
		TableName: jsii.String("glad-entities-" + env),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("EntityType"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("entity_id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		GlobalSecondaryIndexes: &[]*awsdynamodb.GlobalSecondaryIndexPropsV2{
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
		RemovalPolicy:       awscdk.RemovalPolicy_RETAIN, // Keep table on stack deletion
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

	// Export table name and ARN for other stacks
	awscdk.NewCfnOutput(stack, jsii.String("TableName"), &awscdk.CfnOutputProps{
		Value:       entitiesTable.TableName(),
		Description: jsii.String("DynamoDB table name"),
		ExportName:  jsii.String("GladTableName-" + env),
	})

	awscdk.NewCfnOutput(stack, jsii.String("TableArn"), &awscdk.CfnOutputProps{
		Value:       entitiesTable.TableArn(),
		Description: jsii.String("DynamoDB table ARN"),
		ExportName:  jsii.String("GladTableArn-" + env),
	})

	return stack
}
