package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type tableSchema struct {
	name string
	pk   string
	pkT  awsdynamodb.AttributeType
	sk   string
	skT  awsdynamodb.AttributeType
}

func NewF1TelemetryStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	// --- VPC ---
	vpc := awsec2.NewVpc(stack, jsii.String("Vpc"), &awsec2.VpcProps{
		MaxAzs:      jsii.Number(2),
		NatGateways: jsii.Number(0),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
			},
		},
	})

	// --- Security Group ---
	sg := awsec2.NewSecurityGroup(stack, jsii.String("ServerSG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("Allow HTTP/WS on 8080 and SSH"),
		AllowAllOutbound: jsii.Bool(true),
	})
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(8080)), jsii.String("HTTP/WS"), nil)
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(22)), jsii.String("SSH"), nil)

	// --- DynamoDB Tables ---
	tables := []tableSchema{
		{name: "meetings", pk: "MeetingKey", pkT: awsdynamodb.AttributeType_NUMBER},
		{name: "sessions", pk: "MeetingKey", pkT: awsdynamodb.AttributeType_NUMBER, sk: "SessionKey", skT: awsdynamodb.AttributeType_NUMBER},
		{name: "drivers", pk: "SessionKey", pkT: awsdynamodb.AttributeType_NUMBER, sk: "RacingNumber", skT: awsdynamodb.AttributeType_NUMBER},
		{name: "positions", pk: "SessionKey", pkT: awsdynamodb.AttributeType_NUMBER, sk: "RacingNumber", skT: awsdynamodb.AttributeType_NUMBER},
		{name: "timings", pk: "SessionKey", pkT: awsdynamodb.AttributeType_NUMBER, sk: "RacingNumber", skT: awsdynamodb.AttributeType_NUMBER},
		{name: "sectors", pk: "SessionKey", pkT: awsdynamodb.AttributeType_NUMBER, sk: "Reference", skT: awsdynamodb.AttributeType_STRING},
		{name: "trackstatus", pk: "SessionKey", pkT: awsdynamodb.AttributeType_NUMBER, sk: "Utc", skT: awsdynamodb.AttributeType_NUMBER},
		{name: "racecontrol", pk: "SessionKey", pkT: awsdynamodb.AttributeType_NUMBER, sk: "Utc", skT: awsdynamodb.AttributeType_NUMBER},
		{name: "stints", pk: "SessionKey", pkT: awsdynamodb.AttributeType_NUMBER, sk: "RacingNumber", skT: awsdynamodb.AttributeType_NUMBER},
	}

	// --- IAM Role ---
	role := awsiam.NewRole(stack, jsii.String("InstanceRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")),
		},
	})

	var ddbTables []awsdynamodb.Table
	for _, t := range tables {
		tableProps := &awsdynamodb.TableProps{
			TableName:     jsii.String(t.name),
			BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
			PartitionKey: &awsdynamodb.Attribute{
				Name: jsii.String(t.pk),
				Type: t.pkT,
			},
		}
		if t.sk != "" {
			tableProps.SortKey = &awsdynamodb.Attribute{
				Name: jsii.String(t.sk),
				Type: t.skT,
			}
		}
		table := awsdynamodb.NewTable(stack, jsii.String(fmt.Sprintf("Table-%s", t.name)), tableProps)
		table.GrantReadWriteData(role)
		ddbTables = append(ddbTables, table)
	}

	// --- EC2 Instance ---
	userData := awsec2.UserData_ForLinux(&awsec2.LinuxUserDataOptions{})
	userData.AddCommands(
		jsii.String("dnf install -y docker git"),
		jsii.String("systemctl enable docker"),
		jsii.String("systemctl start docker"),
		jsii.String("cd /home/ec2-user"),
		jsii.String("git clone https://github.com/bpougala/F1Telemetry-new-server.git app"),
		jsii.String("cd app"),
		jsii.String("docker build -t f1telemetry ."),
		jsii.String("docker run -d --name f1telemetry --restart unless-stopped -p 8080:8080 -e AWS_DEFAULT_REGION=eu-west-1 f1telemetry"),
	)

	instance := awsec2.NewInstance(stack, jsii.String("Server"), &awsec2.InstanceProps{
		Vpc:           vpc,
		InstanceType:  awsec2.NewInstanceType(jsii.String("t3.small")),
		MachineImage:  awsec2.MachineImage_LatestAmazonLinux2023(&awsec2.AmazonLinux2023ImageSsmParameterProps{}),
		SecurityGroup: sg,
		Role:          role,
		UserData:      userData,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		AssociatePublicIpAddress: jsii.Bool(true),
	})

	// --- Outputs ---
	awscdk.NewCfnOutput(stack, jsii.String("InstancePublicIP"), &awscdk.CfnOutputProps{
		Value:       instance.InstancePublicIp(),
		Description: jsii.String("Public IP of the EC2 instance"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("ServerURL"), &awscdk.CfnOutputProps{
		Value:       jsii.String(fmt.Sprintf("http://%s:8080", *instance.InstancePublicIp())),
		Description: jsii.String("F1 Telemetry server URL"),
	})

	return stack
}
