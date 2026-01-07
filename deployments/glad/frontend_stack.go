package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type FrontendStackProps struct {
	awscdk.StackProps
}

func NewFrontendStack(scope constructs.Construct, id string, props *FrontendStackProps, env string) awscdk.Stack {
	var sprops awscdk.StackProps

	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, &id, &sprops)

	awscdk.Tags_Of(stack).Add(jsii.String("Environment"), jsii.String(env), nil)

	// Create S3 bucket for Angular application hosting
	websiteBucket := awss3.NewBucket(stack, jsii.String(id+"-website-bucket"), &awss3.BucketProps{
		BucketName:        jsii.String("glad-frontend-" + env),
		PublicReadAccess:  jsii.Bool(false), // Private bucket, accessed via CloudFront OAI
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true), // Delete objects when stack is destroyed
		Versioned:         jsii.Bool(false),
	})

	// Create Origin Access Identity for CloudFront to access S3
	oai := awscloudfront.NewOriginAccessIdentity(stack, jsii.String(id+"-oai"), &awscloudfront.OriginAccessIdentityProps{
		Comment: jsii.String("OAI for GLAD Stack frontend " + env),
	})

	// Grant read permissions to CloudFront OAI
	websiteBucket.GrantRead(oai.GrantPrincipal(), jsii.String("*"))

	// Create CloudFront distribution
	distribution := awscloudfront.NewDistribution(stack, jsii.String(id+"-distribution"), &awscloudfront.DistributionProps{
		Comment: jsii.String("GLAD Stack Angular Frontend Distribution - " + env),
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: awscloudfrontorigins.NewS3Origin(websiteBucket, &awscloudfrontorigins.S3OriginProps{
				OriginAccessIdentity: oai,
			}),
			ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
			AllowedMethods:       awscloudfront.AllowedMethods_ALLOW_GET_HEAD_OPTIONS(),
			CachedMethods:        awscloudfront.CachedMethods_CACHE_GET_HEAD_OPTIONS(),
			Compress:             jsii.Bool(true),
			CachePolicy:          awscloudfront.CachePolicy_CACHING_OPTIMIZED(),
		},
		DefaultRootObject: jsii.String("index.html"),
		// SPA routing: redirect 404/403 errors to index.html
		ErrorResponses: &[]*awscloudfront.ErrorResponse{
			{
				HttpStatus:         jsii.Number(404),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
				Ttl:                awscdk.Duration_Seconds(jsii.Number(300)),
			},
			{
				HttpStatus:         jsii.Number(403),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
				Ttl:                awscdk.Duration_Seconds(jsii.Number(300)),
			},
		},
		PriceClass:             awscloudfront.PriceClass_PRICE_CLASS_100, // US, Canada, Europe
		EnableIpv6:             jsii.Bool(true),
		HttpVersion:            awscloudfront.HttpVersion_HTTP2_AND_3,
		EnableLogging:          jsii.Bool(false), // Can enable for production
		MinimumProtocolVersion: awscloudfront.SecurityPolicyProtocol_TLS_V1_2_2021,
	})

	// CloudFormation outputs
	awscdk.NewCfnOutput(stack, jsii.String("WebsiteBucketName"), &awscdk.CfnOutputProps{
		Value:       websiteBucket.BucketName(),
		Description: jsii.String("S3 Bucket for Angular Application"),
		ExportName:  jsii.String("GladWebsiteBucketName-" + env),
	})

	awscdk.NewCfnOutput(stack, jsii.String("WebsiteBucketArn"), &awscdk.CfnOutputProps{
		Value:       websiteBucket.BucketArn(),
		Description: jsii.String("S3 Bucket ARN"),
		ExportName:  jsii.String("GladWebsiteBucketArn-" + env),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DistributionId"), &awscdk.CfnOutputProps{
		Value:       distribution.DistributionId(),
		Description: jsii.String("CloudFront Distribution ID"),
		ExportName:  jsii.String("GladDistributionId-" + env),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DistributionDomainName"), &awscdk.CfnOutputProps{
		Value:       distribution.DistributionDomainName(),
		Description: jsii.String("CloudFront Distribution Domain Name"),
		ExportName:  jsii.String("GladDistributionDomainName-" + env),
	})

	awscdk.NewCfnOutput(stack, jsii.String("WebsiteURL"), &awscdk.CfnOutputProps{
		Value:       jsii.String("https://" + *distribution.DistributionDomainName()),
		Description: jsii.String("Angular Application URL"),
		ExportName:  jsii.String("GladWebsiteURL-" + env),
	})

	return stack
}
