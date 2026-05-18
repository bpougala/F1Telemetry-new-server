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
		Description:      jsii.String("Allow HTTPS, HTTP and SSH"),
		AllowAllOutbound: jsii.Bool(true),
	})
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(80)), jsii.String("HTTP"), nil)
	sg.AddIngressRule(awsec2.Peer_AnyIpv4(), awsec2.Port_Tcp(jsii.Number(443)), jsii.String("HTTPS"), nil)
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
		jsii.String("docker run -d --name f1telemetry --restart unless-stopped -p 8080:8080 -e AWS_DEFAULT_REGION=eu-west-1 f1telemetry"),
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

	return stack
}
