package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

func newTestStack() assertions.Template {
	app := awscdk.NewApp(nil)
	stack := NewF1TelemetryStack(app, "TestStack", &awscdk.StackProps{
		Env: &awscdk.Environment{
			Region: jsii.String("eu-west-1"),
		},
	})
	return assertions.Template_FromStack(stack, nil)
}

func TestDynamoDBTableCount(t *testing.T) {
	template := newTestStack()
	template.ResourceCountIs(jsii.String("AWS::DynamoDB::Table"), jsii.Number(9))
}

func TestAllTablesPayPerRequest(t *testing.T) {
	template := newTestStack()
	template.AllResourcesProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"BillingMode": "PAY_PER_REQUEST",
	})
}

func TestMeetingsTableSchema(t *testing.T) {
	template := newTestStack()
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "meetings",
		"KeySchema": []map[string]interface{}{
			{"AttributeName": "MeetingKey", "KeyType": "HASH"},
		},
		"AttributeDefinitions": []map[string]interface{}{
			{"AttributeName": "MeetingKey", "AttributeType": "N"},
		},
	})
}

func TestSessionsTableSchema(t *testing.T) {
	template := newTestStack()
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "sessions",
		"KeySchema": []map[string]interface{}{
			{"AttributeName": "MeetingKey", "KeyType": "HASH"},
			{"AttributeName": "SessionKey", "KeyType": "RANGE"},
		},
	})
}

func TestSectorsSortKeyIsString(t *testing.T) {
	template := newTestStack()
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), map[string]interface{}{
		"TableName": "sectors",
		"AttributeDefinitions": assertions.Match_ArrayWith(&[]interface{}{
			map[string]interface{}{
				"AttributeName": "Reference",
				"AttributeType": "S",
			},
		}),
	})
}

func TestVPCExists(t *testing.T) {
	template := newTestStack()
	template.ResourceCountIs(jsii.String("AWS::EC2::VPC"), jsii.Number(1))
}

func TestSecurityGroupAllows8080(t *testing.T) {
	template := newTestStack()
	template.HasResourceProperties(jsii.String("AWS::EC2::SecurityGroup"), map[string]interface{}{
		"SecurityGroupIngress": assertions.Match_ArrayWith(&[]interface{}{
			assertions.Match_ObjectLike(&map[string]interface{}{
				"IpProtocol": "tcp",
				"FromPort":   8080,
				"ToPort":     8080,
				"CidrIp":     "0.0.0.0/0",
			}),
		}),
	})
}

func TestEC2InstanceType(t *testing.T) {
	template := newTestStack()
	template.HasResourceProperties(jsii.String("AWS::EC2::Instance"), map[string]interface{}{
		"InstanceType": "t3.small",
	})
}

func TestIAMPolicyGrantsDynamoDB(t *testing.T) {
	template := newTestStack()
	template.HasResourceProperties(jsii.String("AWS::IAM::Policy"), map[string]interface{}{
		"PolicyDocument": map[string]interface{}{
			"Statement": assertions.Match_ArrayWith(&[]interface{}{
				assertions.Match_ObjectLike(&map[string]interface{}{
					"Action": assertions.Match_ArrayWith(&[]interface{}{
						"dynamodb:PutItem",
					}),
				}),
			}),
		},
	})
}

func TestInstanceProfileExists(t *testing.T) {
	template := newTestStack()
	template.ResourceCountIs(jsii.String("AWS::IAM::InstanceProfile"), jsii.Number(1))
}

func TestNoNATGateways(t *testing.T) {
	template := newTestStack()
	template.ResourceCountIs(jsii.String("AWS::EC2::NatGateway"), jsii.Number(0))
}
