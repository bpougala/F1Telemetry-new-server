package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticache"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdanodejs"
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
		Description:      jsii.String("Allow HTTPS, HTTP and SSH"),
		AllowAllOutbound: jsii.Bool(true),
	})
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(80)), jsii.String("HTTP"), nil)
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(443)), jsii.String("HTTPS"), nil)
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(22)), jsii.String("SSH"), nil)

	// --- Valkey Security Group ---
	valkeySG := awsec2.NewSecurityGroup(stack, jsii.String("ValkeySG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("ElastiCache Valkey serverless"),
		AllowAllOutbound: jsii.Bool(false),
	})

	// --- Challenge Lambda Security Group ---
	lambdaSG := awsec2.NewSecurityGroup(stack, jsii.String("ChallengeLambdaSG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("Challenge Lambda function"),
		AllowAllOutbound: jsii.Bool(true),
	})

	valkeySG.AddIngressRule(
		awsec2.Peer_SecurityGroupId(lambdaSG.SecurityGroupId(), nil),
		awsec2.Port_Tcp(jsii.Number(6379)),
		jsii.String("Lambda to Valkey"),
		nil,
	)

	valkeySG.AddIngressRule(
		awsec2.Peer_SecurityGroupId(sg.SecurityGroupId(), nil),
		awsec2.Port_Tcp(jsii.Number(6379)),
		jsii.String("EC2 Server to Valkey"),
		nil,
	)

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
		{name: "weather", pk: "SessionKey", pkT: awsdynamodb.AttributeType_NUMBER, sk: "Utc", skT: awsdynamodb.AttributeType_NUMBER},
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

	// --- ElastiCache Valkey Serverless ---
	publicSubnets := vpc.SelectSubnets(&awsec2.SubnetSelection{
		SubnetType: awsec2.SubnetType_PUBLIC,
	})

	subnetIds := make([]interface{}, len(*publicSubnets.SubnetIds))
	for i, id := range *publicSubnets.SubnetIds {
		subnetIds[i] = id
	}

	valkeyCache := awselasticache.NewCfnServerlessCache(stack, jsii.String("ValkeyCache"), &awselasticache.CfnServerlessCacheProps{
		Engine:              jsii.String("valkey"),
		ServerlessCacheName: jsii.String("f1telemetry-challenges"),
		Description:         jsii.String("Challenge token store"),
		SecurityGroupIds:    &[]interface{}{valkeySG.SecurityGroupId()},
		SubnetIds:           &subnetIds,
		CacheUsageLimits: &awselasticache.CfnServerlessCache_CacheUsageLimitsProperty{
			DataStorage: &awselasticache.CfnServerlessCache_DataStorageProperty{
				Unit:    jsii.String("GB"),
				Maximum: jsii.Number(1),
			},
			EcpuPerSecond: &awselasticache.CfnServerlessCache_ECPUPerSecondProperty{
				Maximum: jsii.Number(1000),
			},
		},
	})
	valkeyCache.ApplyRemovalPolicy(awscdk.RemovalPolicy_DESTROY, &awscdk.RemovalPolicyOptions{})

	// --- Challenge Lambda ---
	challengeFn := awslambdanodejs.NewNodejsFunction(stack, jsii.String("ChallengeFn"), &awslambdanodejs.NodejsFunctionProps{
		Entry:            jsii.String("lambda/challenge/index.ts"),
		Handler:          jsii.String("handler"),
		Runtime:          awslambda.Runtime_NODEJS_22_X(),
		Timeout:          awscdk.Duration_Seconds(jsii.Number(10)),
		DepsLockFilePath: jsii.String("lambda/challenge/package-lock.json"),
		Vpc:              vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		AllowPublicSubnet: jsii.Bool(true),
		SecurityGroups:    &[]awsec2.ISecurityGroup{lambdaSG},
		Environment: &map[string]*string{
			"VALKEY_ENDPOINT": valkeyCache.AttrEndpointAddress(),
			"VALKEY_PORT":     valkeyCache.AttrEndpointPort(),
		},
		Bundling: &awslambdanodejs.BundlingOptions{
			Minify: jsii.Bool(true),
		},
	})

	fnUrl := challengeFn.AddFunctionUrl(&awslambda.FunctionUrlOptions{
		AuthType: awslambda.FunctionUrlAuthType_NONE,
	})

	// --- Verify Lambda ---
	verifyFn := awslambdanodejs.NewNodejsFunction(stack, jsii.String("VerifyFn"), &awslambdanodejs.NodejsFunctionProps{
		Entry:            jsii.String("lambda/verify/index.ts"),
		Handler:          jsii.String("handler"),
		Runtime:          awslambda.Runtime_NODEJS_22_X(),
		Timeout:          awscdk.Duration_Seconds(jsii.Number(10)),
		DepsLockFilePath: jsii.String("lambda/verify/package-lock.json"),
		Vpc:              vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		AllowPublicSubnet: jsii.Bool(true),
		SecurityGroups:    &[]awsec2.ISecurityGroup{lambdaSG},
		Environment: &map[string]*string{
			"VALKEY_ENDPOINT": valkeyCache.AttrEndpointAddress(),
			"VALKEY_PORT":     valkeyCache.AttrEndpointPort(),
			"APP_ID":          jsii.String("pushlap.PUSH-LAP"),
			"DEV_TOKEN":       jsii.String("its-a-secret"),
		},
		Bundling: &awslambdanodejs.BundlingOptions{
			Minify: jsii.Bool(true),
		},
	})

	verifyFnUrl := verifyFn.AddFunctionUrl(&awslambda.FunctionUrlOptions{
		AuthType: awslambda.FunctionUrlAuthType_NONE,
	})

	// --- EC2 Instance ---
	userData := awsec2.UserData_ForLinux(&awsec2.LinuxUserDataOptions{})
	userData.AddCommands(
		// Docker & app
		jsii.String("dnf install -y docker git"),
		jsii.String("systemctl enable docker"),
		jsii.String("systemctl start docker"),
		jsii.String("usermod -aG docker ec2-user"),
		jsii.String("cd /home/ec2-user"),
		jsii.String("git clone https://github.com/bpougala/F1Telemetry-new-server.git app"),
		jsii.String("chown -R ec2-user:ec2-user app"),
		jsii.String("cd app"),
		jsii.String("docker build -t f1telemetry ."),
		awscdk.Fn_Join(jsii.String(""), &[]*string{
			jsii.String("docker run -d --name f1telemetry --restart unless-stopped -p 8080:8080"),
			jsii.String(" -e AWS_DEFAULT_REGION=eu-west-1"),
			jsii.String(" -e VALKEY_ENDPOINT="),
			valkeyCache.AttrEndpointAddress(),
			jsii.String(" -e VALKEY_PORT="),
			valkeyCache.AttrEndpointPort(),
			jsii.String(" f1telemetry"),
		}),
		// Caddy reverse proxy
		jsii.String("curl -fsSL https://github.com/caddyserver/caddy/releases/download/v2.9.1/caddy_2.9.1_linux_amd64.tar.gz -o /tmp/caddy.tar.gz"),
		jsii.String("tar -xzf /tmp/caddy.tar.gz -C /usr/bin caddy"),
		jsii.String("chmod +x /usr/bin/caddy"),
		jsii.String("mkdir -p /etc/caddy"),
		jsii.String("groupadd --system caddy || true"),
		jsii.String("useradd --system --gid caddy --create-home --home /var/lib/caddy --shell /usr/sbin/nologin caddy || true"),
		jsii.String(`cat > /etc/systemd/system/caddy.service <<'UNIT'
[Unit]
Description=Caddy
After=network.target network-online.target
Requires=network-online.target

[Service]
Type=notify
User=caddy
Group=caddy
ExecStart=/usr/bin/caddy run --environ --config /etc/caddy/Caddyfile
ExecReload=/usr/bin/caddy reload --config /etc/caddy/Caddyfile --force
TimeoutStopSec=5s
LimitNOFILE=1048576
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
UNIT`),
		jsii.String("systemctl daemon-reload"),
		jsii.String("cp /home/ec2-user/app/Caddyfile /etc/caddy/Caddyfile"),
		jsii.String("systemctl enable caddy"),
		jsii.String("systemctl start caddy"),
	)

	keyPair := awsec2.KeyPair_FromKeyPairName(stack, jsii.String("KeyPair"), jsii.String("f1telemetry"))

	instance := awsec2.NewInstance(stack, jsii.String("Server"), &awsec2.InstanceProps{
		Vpc:           vpc,
		InstanceType:  awsec2.NewInstanceType(jsii.String("t3.small")),
		MachineImage:  awsec2.MachineImage_LatestAmazonLinux2023(&awsec2.AmazonLinux2023ImageSsmParameterProps{}),
		SecurityGroup: sg,
		Role:          role,
		UserData:      userData,
		KeyPair:       keyPair,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		AssociatePublicIpAddress: jsii.Bool(true),
	})

	// --- Elastic IP ---
	eip := awsec2.NewCfnEIP(stack, jsii.String("ServerEIP"), &awsec2.CfnEIPProps{})
	awsec2.NewCfnEIPAssociation(stack, jsii.String("EIPAssoc"), &awsec2.CfnEIPAssociationProps{
		AllocationId: eip.AttrAllocationId(),
		InstanceId:   instance.InstanceId(),
	})

	// --- Outputs ---
	awscdk.NewCfnOutput(stack, jsii.String("ElasticIP"), &awscdk.CfnOutputProps{
		Value:       eip.AttrPublicIp(),
		Description: jsii.String("Elastic IP — point pushlap.co DNS A record here"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("ServerURL"), &awscdk.CfnOutputProps{
		Value:       jsii.String("https://pushlap.co"),
		Description: jsii.String("F1 Telemetry server URL"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("ChallengeURL"), &awscdk.CfnOutputProps{
		Value:       fnUrl.Url(),
		Description: jsii.String("Challenge Lambda Function URL"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("VerifyURL"), &awscdk.CfnOutputProps{
		Value:       verifyFnUrl.Url(),
		Description: jsii.String("Verify Lambda Function URL"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("ValkeyEndpoint"), &awscdk.CfnOutputProps{
		Value:       valkeyCache.AttrEndpointAddress(),
		Description: jsii.String("ElastiCache Valkey endpoint"),
	})

	return stack
}
